package fsm

import (
	"time"

	"github.com/musenwill/raftdemo/log"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/zap"
)

type Leader struct {
	node  raft.NodeInstance
	nodes []string

	nextIndex  map[string]int64
	matchIndex map[string]int64

	proxy proxy.Proxy
	cfg   *raft.Config

	leaving chan bool

	logger log.Logger
}

func NewLeader(node raft.NodeInstance, nodes []string, proxy proxy.Proxy, cfg *raft.Config, logger log.Logger) *Leader {
	leader := &Leader{
		node:    node,
		nodes:   nodes,
		proxy:   proxy,
		cfg:     cfg,
		leaving: make(chan bool),
		logger:  *logger.With(zap.String("state", "leader")),

		nextIndex:  make(map[string]int64),
		matchIndex: make(map[string]int64),
	}

	lastLogIndex := node.GetLastLogIndex()
	for _, n := range nodes {
		leader.nextIndex[n] = lastLogIndex
	}

	return leader
}

func (s *Leader) Enter() {
	s.node.SetReadable(false)
	s.waitApply()
	s.notifyWinVote()
	s.node.AppendNop()

	go func() {
		ticker := time.NewTicker(s.cfg.ReplicateTimeout)
		for {
			select {
			case <-s.leaving:
				return
			case <-ticker.C:
				s.OnTimeout()
			}
		}
	}()
}

func (s *Leader) Leave() {
	close(s.leaving)
	s.node.SetReadable(true)
}

func (s *Leader) State() model.StateRole {
	return model.StateRole_Leader
}

func (s *Leader) OnAppendEntries(param model.AppendEntries) model.Response {
	if param.Term > s.node.GetTerm() {
		s.node.SwitchStateTo(model.StateRole_Follower)
		return s.node.OnAppendEntries(param)
	}

	// leader reject any append entries
	return model.Response{Term: s.node.GetTerm(), Success: false}
}

func (s *Leader) OnRequestVote(param model.RequestVote) model.Response {
	if param.Term > s.node.GetTerm() {
		s.node.SwitchStateTo(model.StateRole_Follower)
		return s.node.OnRequestVote(param)
	}

	return model.Response{Term: s.node.GetTerm(), Success: false}
}

func (s *Leader) notifyWinVote() {
	go func() {
		for _, n := range s.nodes {
			nodeID := n
			select {
			case <-s.leaving:
				break
			default:
			}
			go func() {
				response, err := s.proxy.Send(nodeID, model.AppendEntries{
					Term:         s.node.GetTerm(),
					PrevLogIndex: s.node.GetLastLogIndex(),
					PrevLogTerm:  s.node.GetLastLogTerm(),
					LeaderCommit: s.node.GetCommitIndex(),
					LeaderID:     s.node.GetNodeID(),
				}, s.leaving)
				if err != nil {
					s.logger.Errorf("notify win vote, %w", err)
					return
				}
				if response.Term > s.node.GetTerm() {
					s.node.SwitchStateTo(model.StateRole_Follower)
				}
			}()
		}
	}()
}

func (s *Leader) waitApply() {
	go func() {
		s.node.WaitApply(s.leaving)
		s.node.SetReadable(true)
	}()
}

func (s *Leader) OnTimeout() {
	go func() {
		for _, n := range s.nodes {
			nodeID := n
			select {
			case <-s.leaving:
				break
			default:
			}
			go func() {
				nextIndex := s.nextIndex[nodeID]
				if nextIndex == 0 {
					nextIndex = 1
				}
				entries := s.node.GetFollowingEntries(nextIndex)
				preEntry, err := s.node.GetEntry(nextIndex - 1)
				if err != nil {
					s.logger.Errorf("append entries, %w", err)
					return
				}

				response, err := s.proxy.Send(nodeID, model.AppendEntries{
					Term:         s.node.GetTerm(),
					PrevLogIndex: preEntry.Id,
					PrevLogTerm:  preEntry.Term,
					LeaderCommit: s.node.GetCommitIndex(),
					LeaderID:     s.node.GetNodeID(),
					Entries:      entries,
				}, s.leaving)
				if err != nil {
					s.logger.Errorf("append entries, %w", err)
					return
				}
				if response.Term > s.node.GetTerm() {
					s.node.SwitchStateTo(model.StateRole_Follower)
				}
				if !response.Success {
					nextIndex--
					if nextIndex < 0 {
						nextIndex = 0
					}
					s.nextIndex[nodeID] = nextIndex
					s.logger.Info("failed replica log", zap.String("nodeID", nodeID), zap.Int64("logID", nextIndex))
				} else {
					s.logger.Info("success replica log", zap.String("nodeID", nodeID), zap.Int64("logID", nextIndex))
				}
			}()
		}
	}()
}

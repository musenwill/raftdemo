package fsm

import (
	"sync"
	"time"

	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/zap"
)

type Leader struct {
	node raft.NodeInstance

	muNextIndex struct {
		sync.RWMutex
		m map[string]int64
	}
	muMatchIndex struct {
		sync.RWMutex
		m map[string]int64
	}
	nodes []string

	cfg     *raft.Config
	leaving chan bool
	logger  *zap.Logger
}

func NewLeader(node raft.NodeInstance, nodes []string, cfg *raft.Config, logger *zap.Logger) *Leader {
	leader := &Leader{
		node:    node,
		cfg:     cfg,
		leaving: make(chan bool),
		logger:  logger.With(zap.String("state", "leader")),

		nodes: nodes,
	}
	leader.muNextIndex.m = make(map[string]int64)
	leader.muMatchIndex.m = make(map[string]int64)

	lastLogIndex := node.GetLastEntry().Id
	for _, n := range nodes {
		leader.muNextIndex.m[n] = lastLogIndex + 1
		leader.muMatchIndex.m[n] = 0
	}

	return leader
}

func (s *Leader) Enter() {
	s.printLog(s.logger.Info, "enter state")

	s.node.ResetLeader()
	s.node.ResetVoteFor()
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
	s.printLog(s.logger.Info, "leave state")
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
	term := s.node.GetTerm()
	lastEntry := s.node.GetLastEntry()
	request := &model.AppendEntries{
		Term:         s.node.GetTerm(),
		PrevLogIndex: lastEntry.Id,
		PrevLogTerm:  lastEntry.Term,
		LeaderCommit: s.node.GetCommitIndex(),
		LeaderID:     s.node.GetNodeID(),
	}

	getRequestF := func(nodeID string) (interface{}, error) {
		return *request, nil
	}
	handleResponseF := func(nodeID string, response model.Response) {
		if response.Term > term {
			s.node.SwitchStateTo(model.StateRole_Follower)
		}
	}

	s.node.Broadcast("win vote", s.leaving, getRequestF, handleResponseF)
}

func (s *Leader) waitApply() {
	go func() {
		s.node.WaitApply(s.leaving)
		s.node.SetReadable(true)
	}()
}

func (s *Leader) OnTimeout() {
	s.printLog(s.logger.Debug, "start append entries")
	s.node.Broadcast("append request", s.leaving, s.getRequest, s.handleResponse)
	s.CheckMatchIndex()
}

func (s *Leader) getRequest(nodeID string) (interface{}, error) {
	s.muNextIndex.RLock()
	nextIndex := s.muNextIndex.m[nodeID]
	s.muNextIndex.RUnlock()

	if nextIndex <= 0 {
		nextIndex = 1
	}
	entries := s.node.GetFollowingEntries(nextIndex)
	preEntry, err := s.node.GetEntry(nextIndex - 1)
	if err != nil {
		return model.AppendEntries{}, err
	}

	return model.AppendEntries{
		Term:         s.node.GetTerm(),
		PrevLogIndex: preEntry.Id,
		PrevLogTerm:  preEntry.Term,
		LeaderCommit: s.node.GetCommitIndex(),
		LeaderID:     s.node.GetNodeID(),
		Entries:      entries,
	}, nil
}

func (s *Leader) handleResponse(nodeID string, response model.Response) {
	if response.Term > s.node.GetTerm() {
		s.node.SwitchStateTo(model.StateRole_Follower)
		return
	}

	lastEntry := s.node.GetLastEntry()

	s.muNextIndex.RLock()
	nextIndex := s.muNextIndex.m[nodeID]
	s.muNextIndex.RUnlock()

	if !response.Success {
		nextIndex--
		if nextIndex <= 0 {
			nextIndex = 1
		}
		s.muNextIndex.Lock()
		s.muNextIndex.m[nodeID] = nextIndex
		s.muNextIndex.Unlock()
		s.printLog(s.logger.Debug, "failed replica log", zap.String("peer", nodeID), zap.Int64("lag", lastEntry.Id-nextIndex))
	} else {
		s.muNextIndex.Lock()
		s.muNextIndex.m[nodeID] = lastEntry.Id + 1
		s.muNextIndex.Unlock()

		s.muMatchIndex.Lock()
		s.muMatchIndex.m[nodeID] = lastEntry.Id
		s.muMatchIndex.Unlock()

		s.printLog(s.logger.Debug, "success replica log", zap.String("peer", nodeID), zap.Int64("lag", lastEntry.Id-nextIndex))
	}
}

func (s *Leader) CheckMatchIndex() {
	latestIndexes := make([]int64, 0)

	commitIndex := s.node.GetCommitIndex()
	lastIndex := s.node.GetLastEntry().Id

	if lastIndex > commitIndex {
		latestIndexes = append(latestIndexes, lastIndex)
	}

	s.muMatchIndex.RLock()
	for _, m := range s.muMatchIndex.m {
		if m > commitIndex {
			latestIndexes = append(latestIndexes, m)
		}
	}
	s.muMatchIndex.RUnlock()

	if len(latestIndexes) > (len(s.nodes)+1)/2 {
		min := latestIndexes[0]
		for _, i := range latestIndexes {
			if i < min {
				min = i
			}
		}
		s.node.CASCommitID(min)
	}
}

func (s *Leader) printLog(fn func(msg string, fields ...zap.Field), msg string, fields ...zap.Field) {
	fields = append(fields, zap.Int64("term", s.node.GetTerm()),
		zap.Int64("commitID", s.node.GetCommitIndex()),
		zap.Int64("appliedID", s.node.GetLastAppliedIndex()),
		zap.Bool("readable", s.node.Readable()))
	fn(msg, fields...)
}

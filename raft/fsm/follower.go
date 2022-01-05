package fsm

import (
	"time"

	"github.com/musenwill/raftdemo/log"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/zap"
)

type Follower struct {
	node raft.NodeInstance
	cfg  *raft.Config

	leaving   chan bool
	heartBeat chan bool

	logger log.Logger
}

func NewFollower(node raft.NodeInstance, cfg *raft.Config, logger log.Logger) *Follower {
	return &Follower{
		node:      node,
		cfg:       cfg,
		leaving:   make(chan bool),
		heartBeat: make(chan bool),
		logger:    *logger.With(zap.String("state", "follower")),
	}
}

func (f *Follower) Enter() {
	go func() {
		ticker := time.NewTicker(f.cfg.ElectionTimeout)
		for {
			select {
			case <-f.leaving:
				return
			case <-ticker.C:
				f.OnTimeout()
			case <-f.heartBeat:
				ticker.Reset(f.cfg.ElectionTimeout)
			}
		}
	}()
}

func (f *Follower) Leave() {
	close(f.leaving)
	close(f.heartBeat)
}

func (f *Follower) OnAppendEntries(param model.AppendEntries) model.Response {
	if f.node.CASTerm(param.Term) < 0 {
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	f.heartBeat <- true

	f.node.SetLeader(param.LeaderID)
	entry, err := f.node.GetEntry(param.PrevLogIndex)
	if err != nil {
		f.logger.Error(err)
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}
	if entry.Id == param.PrevLogIndex && entry.Term != param.PrevLogTerm {
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	if err := f.node.AppendEntries(param.Entries); err != nil {
		f.logger.Error("append entries", zap.Int64("preLogIndex", param.PrevLogIndex), zap.Int64("preLogTerm", param.PrevLogTerm))
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	return model.Response{Term: f.node.GetTerm(), Success: true}
}

func (f *Follower) OnRequestVote(param model.RequestVote) model.Response {
	if f.node.CASTerm(param.Term) < 0 {
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	f.heartBeat <- true

	if f.node.GetVoteFor() != "" {
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	lastEntry := f.node.GetLastEntry()

	if lastEntry.Term <= param.LastLogTerm && lastEntry.Id <= param.LastLogIndex {
		f.node.SetVoteFor(param.CandidateID)
		return model.Response{Term: f.node.GetTerm(), Success: true}
	}

	return model.Response{Term: f.node.GetTerm(), Success: false}
}

func (f *Follower) OnTimeout() {
	f.node.SwitchStateTo(model.StateRole_Candidate)
}

func (f *Follower) State() model.StateRole {
	return model.StateRole_Follower
}

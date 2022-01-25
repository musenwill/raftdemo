package fsm

import (
	"time"

	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/zap"
)

type Follower struct {
	node raft.NodeInstance
	cfg  *raft.Config

	leaving   chan bool
	heartBeat chan bool

	logger *zap.Logger
}

func NewFollower(node raft.NodeInstance, cfg *raft.Config, logger *zap.Logger) *Follower {
	return &Follower{
		node:      node,
		cfg:       cfg,
		leaving:   make(chan bool),
		heartBeat: make(chan bool),
		logger:    logger.With(zap.String("state", "follower")),
	}
}

func (f *Follower) Enter() {
	f.printLog(f.logger.Info, "enter state")

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
	f.printLog(f.logger.Info, "leave state")
	close(f.leaving)
	close(f.heartBeat)
}

func (f *Follower) OnAppendEntries(param model.AppendEntries) model.Response {
	if f.node.CASTerm(param.Term) < 0 {
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	select {
	case f.heartBeat <- true:
	default:
	}

	f.node.SetLeader(param.LeaderID)
	entry, err := f.node.GetEntry(param.PrevLogIndex)
	if err != nil {
		f.printLog(f.logger.Error, err.Error())
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}
	if entry.Id == param.PrevLogIndex && entry.Term != param.PrevLogTerm {
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	if err := f.node.AppendEntries(param.Entries); err != nil {
		f.printLog(f.logger.Error, "append entries", zap.Int64("preLogIndex", param.PrevLogIndex), zap.Int64("preLogTerm", param.PrevLogTerm))
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	f.node.CASCommitID(param.LeaderCommit)

	return model.Response{Term: f.node.GetTerm(), Success: true}
}

func (f *Follower) OnRequestVote(param model.RequestVote) model.Response {
	if f.node.CASTerm(param.Term) < 0 {
		return model.Response{Term: f.node.GetTerm(), Success: false}
	}

	select {
	case f.heartBeat <- true:
	default:
	}

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
	f.node.SwitchStateTo(model.StateRole_candidate)
}

func (f *Follower) State() model.StateRole {
	return model.StateRole_follower
}

func (f *Follower) printLog(fn func(msg string, fields ...zap.Field), msg string, fields ...zap.Field) {
	fields = append(fields, zap.Int64("term", f.node.GetTerm()),
		zap.Int64("commitID", f.node.GetCommitIndex()),
		zap.Int64("appliedID", f.node.GetLastAppliedIndex()),
		zap.Bool("readable", f.node.Readable()))
	fn(msg, fields...)
}

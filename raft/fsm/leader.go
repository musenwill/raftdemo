package fsm

import (
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

	logger log.Logger
}

func NewLeader(node raft.NodeInstance, nodes []string, proxy proxy.Proxy, cfg *raft.Config, logger log.Logger) *Leader {
	return &Leader{
		node:   node,
		nodes:  nodes,
		proxy:  proxy,
		cfg:    cfg,
		logger: *logger.With(zap.String("state", "leader")),

		nextIndex:  make(map[string]int64),
		matchIndex: make(map[string]int64),
	}
}

func (s *Leader) State() model.StateRole {
	return model.StateRole_Leader
}

func (s *Leader) OnAppendEntries(request model.AppendEntries) model.Response {
	if request.Term > s.node.GetTerm() {
		s.node.SwitchStateTo(model.StateRole_Follower)
		return s.node.OnAppendEntries(request)
	}

	// leader reject any append entries
	return model.Response{Term: s.node.GetTerm(), Success: false}
}

func (s *Leader) OnRequestVote(request model.RequestVote) model.Response {
	if request.Term > s.node.GetTerm() {
		s.node.SwitchStateTo(model.StateRole_Follower)
		return s.node.OnRequestVote(request)
	}

	return model.Response{Term: s.node.GetTerm(), Success: false}
}

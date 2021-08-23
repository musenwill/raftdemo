package fsm

import (
	"github.com/musenwill/raftdemo/log"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
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

func (s *Leader) OnAppendEntries(request model.AppendEntries) model.Response {
	s.node.SetTerm(request.Term)
	s.node.SetCommitIndex(request.LeaderCommit)
	s.node.SetLeader(request.LeaderID)

	return model.Response{}
}

func (s *Leader) OnRequestVote(request model.RequestVote) model.Response {
	s.node.SetTerm(request.Term)

	return model.Response{}
}

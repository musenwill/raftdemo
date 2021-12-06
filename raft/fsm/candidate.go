package fsm

import (
	"github.com/musenwill/raftdemo/log"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/zap"
)

type Candidate struct {
	node raft.NodeInstance
	cfg  *raft.Config

	leaving chan bool

	nodes []string
	proxy proxy.Proxy

	logger log.Logger
}

func NewCandidate(node raft.NodeInstance, nodes []string, proxy proxy.Proxy, cfg *raft.Config, logger log.Logger) *Candidate {
	return &Candidate{
		node:    node,
		nodes:   nodes,
		proxy:   proxy,
		cfg:     cfg,
		leaving: make(chan bool),
		logger:  *logger.With(zap.String("state", "follower")),
	}
}

func (c *Candidate) Enter() {

}

func (c *Candidate) Leave() {

}

func (c *Candidate) OnAppendEntries(param model.AppendEntries) model.Response {

}

func (c *Candidate) OnRequestVote(param model.RequestVote) model.Response {

}

func (c *Candidate) OnTimeout() {

}

func (c *Candidate) State() model.StateRole {

}

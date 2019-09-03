package fsm

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
)

type nodeIndex struct {
	nextIndex, matchIndex int64
}

type Leader struct {
	*Server // embed server
	nIndex  map[string]nodeIndex
}

func NewLeader(s *Server, conf *config.Config) *Leader {
	nIndex := make(map[string]nodeIndex)
	for _, n := range conf.Nodes {
		nIndex[n.ID] = nodeIndex{}
	}
	return &Leader{s, nIndex}
}

func (p *Leader) implStateInterface() {
	var _ State = &Leader{}
}

func (p *Leader) enterState() {
}

func (p *Leader) leaveState() {
}

func (p *Leader) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	return proxy.Response{}
}

func (p *Leader) onRequestVote(param proxy.RequestVote) proxy.Response {
	return proxy.Response{}
}

func (p *Leader) timeout() {

}

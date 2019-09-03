package fsm

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
)

type nodeIndex struct {
	nextIndex, matchIndex int64
}

type Leader struct {
	*Server       // embed server
	nIndex        map[string]nodeIndex
	stopReplicate chan bool
}

func NewLeader(s *Server, conf *config.Config) *Leader {
	nIndex := make(map[string]nodeIndex)
	for _, n := range conf.Nodes {
		nIndex[n.ID] = nodeIndex{}
	}
	return &Leader{s, nIndex, nil}
}

func (p *Leader) implStateInterface() {
	var _ State = &Leader{}
}

func (p *Leader) enterState() {
	p.resetTimer()
	if p.stopReplicate != nil {
		close(p.stopReplicate)
	}
	p.stopReplicate = nil
}

func (p *Leader) leaveState() {
	if p.stopReplicate != nil {
		close(p.stopReplicate)
	}
	p.stopReplicate = nil
}

func (p *Leader) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	if param.Term <= p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.onAppendEntries(param)
	}
}

func (p *Leader) onRequestVote(param proxy.RequestVote) proxy.Response {
	if param.Term <= p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.onRequestVote(param)
	}
}

func (p *Leader) timeout() {
	p.resetTimer()
}

package fsm

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type Dummy struct {
	stateLogger *zap.SugaredLogger
}

func NewDummyState(s *Server, config *config.Config) *Dummy {
	return &Dummy{s.logger.With("state", "candidate")}
}

func (p *Dummy) implStateInterface() {
	var _ State = &Dummy{}
}

func (p *Dummy) enterState() {
	p.stateLogger.Info("enter state")
}

func (p *Dummy) leaveState() {
	p.stateLogger.Info("leave state")
}

func (p *Dummy) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	return proxy.Response{}
}

func (p *Dummy) onRequestVote(param proxy.RequestVote) proxy.Response {
	return proxy.Response{}
}

func (p *Dummy) timeout() {
}

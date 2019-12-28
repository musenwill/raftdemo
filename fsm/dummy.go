package fsm

import (
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type Dummy struct {
	stateLogger *zap.SugaredLogger
}

func NewDummyState(s Loggable) *Dummy {
	return &Dummy{s.GetLogger().With("state", StateEnum.None)}
}

func (p *Dummy) implStateInterface() {
	var _ State = &Dummy{}
}

func (p *Dummy) GetLogger() *zap.SugaredLogger {
	return p.stateLogger
}

func (p *Dummy) EnterState() {
	p.stateLogger.Info("enter state")
}

func (p *Dummy) LeaveState() {
	p.stateLogger.Info("leave state")
}

func (p *Dummy) OnAppendEntries(param proxy.AppendEntries) proxy.Response {
	return proxy.Response{}
}

func (p *Dummy) OnRequestVote(param proxy.RequestVote) proxy.Response {
	return proxy.Response{}
}

func (p *Dummy) Timeout() {
}

func (p *Dummy) GetLeader() string {
	return ""
}

func (p *Dummy) GetVoteFor() string {
	return ""
}

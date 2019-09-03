package fsm

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
)

type State interface {
	enterState()
	leaveState()
	onAppendEntries(param proxy.AppendEntries) proxy.Response
	onRequestVote(param proxy.RequestVote) proxy.Response
	timeout()
}

type Server struct {
	id       string
	votedFor string

	currentTerm int64
	commitIndex int64
	lastAplied  int64
	timer       *time.Timer
	logs        []*proxy.Log

	config *config.Config

	currentState State
}

func NewServer(id string, config *config.Config) *Server {
	s := &Server{
		id:     id,
		logs:   make([]*proxy.Log, 0),
		config: config,
	}

	s.checkConfig(config)

	// initial state is follower
	s.transferState(NewFollower(s, config))
	return s
}

func (p *Server) transferState(state State) {
	if p.currentState != nil {
		p.currentState.leaveState()
	}
	p.currentState = state
	p.currentState.enterState()
}

func (p *Server) resetTimer() {
	if p.timer == nil {
		p.timer = time.NewTimer(time.Duration(p.config.Timeout) * time.Millisecond)
	} else {
		p.timer.Reset(time.Duration(p.config.Timeout) * time.Millisecond)
	}
}

func (p *Server) randomResetTimer() {
	randTime := int(float64(p.config.Timeout) * 1.5)
	if p.timer == nil {
		p.timer = time.NewTimer(time.Duration(randTime) * time.Millisecond)
	} else {
		p.timer.Reset(time.Duration(randTime) * time.Millisecond)
	}
}

func (p *Server) checkConfig(config *config.Config) {
	if config.Timeout <= 0 {
		panic(fmt.Sprintf("invalid timeout %v, expected greater than 0", config.Timeout))
	}
	if config.Timeout < 20 {
		fmt.Printf("timeout too small %v, may cause a lot of election, expected greater than 20ms", config.Timeout)
	}
	if config.Timeout > 1000 {
		fmt.Printf("timeout is %v, may cause performance quite slow, are you sure", config.Timeout)
	}
	if len(config.Nodes) <= 0 {
		panic("empty nodes")
	}
	set := make(map[string]bool)
	for _, id := range config.Nodes {
		if _, ok := set[id.ID]; ok {
			panic("duplicate nodes")
		} else {
			set[id.ID] = true
		}
	}
}

func (p *Server) getCommitIndex() int64 {
	return atomic.LoadInt64(&p.commitIndex)
}

func (p *Server) setCommitIndex(i int64) {
	atomic.StoreInt64(&p.commitIndex, i)
}

func (p *Server) lastLogIndex() int64 {
	return int64(len(p.logs) - 1)
}

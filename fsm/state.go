package fsm

import (
	"fmt"
	"time"
)

type State interface {
	enterState()
	onAppendEntries(param AppendEntries) Response
	onRequestVote(param RequestVote) Response
	timeout()
}

type Server struct {
	id       string
	votedFor string

	currentTerm int
	commitIndex int
	lastAplied  int
	timer       *time.Timer
	logs        []*Log

	config Config

	currentState   State
	followerState  State
	candidateState State
	leaderState    State
}

func NewServer(id string, config Config) *Server {
	s := &Server{
		id:     id,
		logs:   make([]*Log, 0),
		config: config,
	}

	s.checkConfig(&config)

	followerState := NewFollower(s, config)
	candidateState := NewCandidate(s, config)
	leaderState := NewLeader(s, config)

	s.followerState = followerState
	s.candidateState = candidateState
	s.leaderState = leaderState

	// initial state is follower
	s.currentState = s.followerState
	s.currentState.enterState()

	return s
}

func (p *Server) resetTimer() {
	if p.timer == nil {
		p.timer = time.NewTimer(time.Duration(p.config.Timeout) * time.Millisecond)
	} else {
		p.timer.Reset(time.Duration(p.config.Timeout) * time.Millisecond)
	}
}

func (p *Server) randomResetTimer() {

}

func (p *Server) checkConfig(config *Config) {
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
		if ok := set[id.ID]; ok {
			panic("duplicate nodes")
		} else {
			set[id.ID] = true
		}
	}
}

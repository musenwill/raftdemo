package fsm

import (
	"context"
	"sync"

	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
)

type Candidate struct {
	*Server      // embed server
	stopElection chan bool
}

func NewCandidate(s *Server, config *config.Config) *Candidate {
	return &Candidate{s, nil}
}

func (p *Candidate) implStateInterface() {
	var _ State = &Candidate{}
}

func (p *Candidate) enterState() {
	p.randomResetTimer()
	if p.stopElection != nil {
		close(p.stopElection)
	}
	p.stopElection = nil
}

func (p *Candidate) leaveState() {
	if p.stopElection != nil {
		close(p.stopElection)
	}
	p.stopElection = nil
}

func (p *Candidate) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	if param.Term < p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.currentState.onAppendEntries(param)
	}
}

func (p *Candidate) onRequestVote(param proxy.RequestVote) proxy.Response {
	if param.Term <= p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.currentState.onRequestVote(param)
	}
}

func (p *Candidate) timeout() {
	p.currentState.enterState()

	p.currentTerm += 1
	p.votedFor = p.id
	p.stopElection = make(chan bool)
	vote := make(chan bool)

	go p.countVote(vote)
	go p.canvass(vote)
}

func (p *Candidate) countVote(vote chan bool) {
	count := 1 // 1 for self vote
	for {
		select {
		case <-p.stopElection:
			return
		case _, ok := <-vote:
			if !ok {
				return
			}
			count += 1

			// win the election
			if count > len(p.config.Nodes)/2 {
				p.transferState(NewLeader(p.Server, p.config))
				return
			}
		}
	}
}

// @TODO: to be optimized
func (p *Candidate) canvass(vote chan bool) {
	defer close(vote)

	ctx, cancel := context.WithCancel(context.Background())
	// make sure jobs are canceled when election timeout
	go func() {
		<-p.stopElection
		cancel()
	}()

	wg := &sync.WaitGroup{}
	wg.Add(len(p.config.Nodes))

	for range p.config.Nodes {
		select {
		case <-p.stopElection:
			return
		default:
			go func() {
				defer wg.Done()

				response, err := proxy.SendRequestVote(ctx, p.id, proxy.RequestVote{})
				if err != nil {
					// log it
					return
				}
				if response.Term > p.currentTerm {
					p.transferState(NewFollower(p.Server, p.config))
					return
				}
				if response.Success {
					vote <- true
				}
			}()
		}
	}

	wg.Wait()
}

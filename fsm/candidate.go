package fsm

import (
	"sync"

	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type Candidate struct {
	*Server      // embed server
	stopElection chan bool
	stateLogger  *zap.SugaredLogger
}

func NewCandidate(s *Server, config *config.Config) *Candidate {
	return &Candidate{s, nil, s.logger.With("state", "candidate")}
}

func (p *Candidate) implStateInterface() {
	var _ State = &Candidate{}
}

func (p *Candidate) enterState() {
	p.stateLogger.Info("enter state")
	p.randomResetTimer()
	if p.stopElection != nil {
		close(p.stopElection)
	}
	p.stopElection = make(chan bool)
}

func (p *Candidate) leaveState() {
	if p.stopElection != nil {
		close(p.stopElection)
	}
	p.stopElection = nil
	p.stateLogger.Info("leave state")
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

	vote := make(chan bool)
	go p.countVote(vote)
	go p.canvass(vote)
}

func (p *Candidate) countVote(vote chan bool) {
	count := 1 // add self in firstly
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

func (p *Candidate) canvass(vote chan bool) {
	defer close(vote)

	wg := &sync.WaitGroup{}
	wg.Add(len(p.config.Nodes))

	for _, n := range p.config.Nodes {
		node := n
		go func() {
			wg.Done()

			var lastLogTerm int64 = 0
			lastLogIndex := p.lastLogIndex()
			if lastLogIndex >= 0 {
				lastLogTerm = p.logs[lastLogIndex].Term
			}
			request := proxy.RequestVote{
				Term:         p.currentTerm,
				CandidateID:  p.id,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm}

			select {
			case <-p.stopElection:
				return
			case proxy.RequestVoteRequestSender(node.ID) <- request:
				select {
				case <-p.stopElection:
					return
				case response := <-proxy.RequestVoteResponseReader(node.ID):
					if response.Term > p.currentTerm {
						p.transferState(NewFollower(p.Server, p.config))
						return
					}
					if response.Success {
						vote <- true
					}
				}
			}
		}()
	}

	wg.Wait()
}

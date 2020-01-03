package fsm

import (
	"fmt"
	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/proxy"
	"math/rand"
)

type Candidate struct {
	Prober
	stopElection chan bool
	stateLogger  *common.Logger
}

func NewCandidate(s Prober, logger *common.Logger) *Candidate {
	return &Candidate{s, nil, logger.With("state", StateEnum.Candidate)}
}

func (p *Candidate) implStateInterface() {
	var _ State = &Candidate{}
}

func (p *Candidate) GetLogger() *common.Logger {
	return p.stateLogger
}

func (p *Candidate) EnterState() {
	p.stateLogger.Infow("enter state", "term", p.GetTerm(),
		"commitIndex", p.GetCommitIndex(), "lastApplied", p.GetLastAppliedIndex())
	p.randomResetTimer()
}

func (p *Candidate) LeaveState() {
	p.resetStopElection()
	p.stateLogger.Infow("leave state", "term", p.GetTerm(),
		"commitIndex", p.GetCommitIndex(), "lastApplied", p.GetLastAppliedIndex())
}

func (p *Candidate) OnAppendEntries(param proxy.AppendEntries) proxy.Response {
	if param.Term < p.GetTerm() {
		return proxy.Response{Term: p.GetTerm(), Success: false}
	} else {
		p.TransferState(StateEnum.Follower)
		return p.GetCurrentState().OnAppendEntries(param)
	}
}

func (p *Candidate) OnRequestVote(param proxy.RequestVote) proxy.Response {
	if param.Term <= p.GetTerm() {
		return proxy.Response{Term: p.GetTerm(), Success: false}
	} else {
		p.TransferState(StateEnum.Follower)
		return p.GetCurrentState().OnRequestVote(param)
	}
}

func (p *Candidate) Timeout() {
	p.IncreaseTerm()
	p.EnterState()

	vote := make(chan bool)
	go p.countVote(vote)
	go p.canvassJob(vote)
}

func (p *Candidate) GetLeader() string {
	return ""
}

func (p *Candidate) GetVoteFor() string {
	return ""
}

func (p *Candidate) countVote(vote <-chan bool) {
	count := 0
	defer func() {
		p.stateLogger.Infow("finish vote", "vote count", count)
	}()

	for {
		select {
		case <-p.stopElection:
			return
		case _, ok := <-vote:
			if !ok {
				return
			}
			count++

			// win the election
			if count > p.GetConfig().GetNodeCount()/2 {
				p.TransferState(StateEnum.Leader)
				return
			}
		}
	}
}

func (p *Candidate) canvassJob(vote chan<- bool) {
	// give self a vote firstly
	vote <- true
	for _, n := range p.GetConfig().GetNodes() {
		nodeID := n.ID
		if nodeID == p.GetHost() {
			continue
		}

		select {
		case <-p.stopElection:
			return
		default:
			go func() {
				p.canvass(nodeID, vote)
			}()
		}
	}
}

func (p *Candidate) canvass(nodeID string, vote chan<- bool) {
	var lastLogTerm int64 = 0
	lastLogIndex := p.GetLastLogIndex()
	if lastLogIndex >= 0 {
		lastLogTerm = p.GetLogs()[lastLogIndex].Term
	}
	request := proxy.RequestVote{
		Term:         p.GetTerm(),
		CandidateID:  p.GetHost(),
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm}

	voteRequestSender, err := p.GetProxy().VoteRequestSender(nodeID)
	if err != nil {
		p.stateLogger.Error(err)
		return
	}
	voteResponseReader, err := p.GetProxy().VoteResponseReader(nodeID)
	if err != nil {
		p.stateLogger.Error(err)
		return
	}

	select {
	case <-p.stopElection:
		return
	case voteRequestSender <- request:
		select {
		case <-p.stopElection:
			return
		case response, ok := <-voteResponseReader:
			if !ok {
				p.stateLogger.Error(fmt.Sprintf("vote response channel of node %s closed", nodeID))
				return
			}
			if response.Term > p.GetTerm() {
				p.TransferState(StateEnum.Follower)
				return
			}
			if response.Success {
				vote <- true
			}
		}
	}
}

// reset timer for candidate randomly
func (p *Candidate) randomResetTimer() {
	confTimeout := p.GetConfig().GetReplicateTimeout()
	randTime := rand.Int63()%confTimeout + confTimeout/2
	p.resetStopElection()
	_ = p.SetTimer(randTime)
}

func (p *Candidate) resetStopElection() {
	if p.stopElection != nil {
		close(p.stopElection)
	}
	p.stopElection = make(chan bool)
}

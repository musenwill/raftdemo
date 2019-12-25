package fsm

import (
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type Follower struct {
	Prober
	votedFor    string
	leaderID    string
	stateLogger *zap.SugaredLogger
}

func NewFollower(s Prober, logger *zap.SugaredLogger) *Follower {
	return &Follower{s, "", "", logger.With("state", StateEnum.Follower)}
}

func (p *Follower) implStateInterface() {
	var _ State = &Follower{}
}

func (p *Follower) GetLogger() *zap.SugaredLogger {
	return p.stateLogger
}

func (p *Follower) EnterState() {
	p.stateLogger.Infow("enter state", "term", p.GetTerm(), "voteFor", p.votedFor,
		"commitIndex", p.GetCommitIndex(), "lastApplied", p.GetLastAppliedIndex())
	p.resetTimer()
	p.votedFor = ""
}

func (p *Follower) LeaveState() {
	p.stateLogger.Infow("leave state", "term", p.GetTerm(), "voteFor", p.votedFor,
		"commitIndex", p.GetCommitIndex(), "lastApplied", p.GetLastAppliedIndex())
}

func (p *Follower) OnAppendEntries(param proxy.AppendEntries) proxy.Response {
	p.resetTimer()

	// 1. reply false if term < currentTerm
	if param.Term < p.GetTerm() {
		return proxy.Response{Term: p.GetTerm(), Success: false}
	}
	p.SetTerm(param.Term)
	p.leaderID = param.LeaderID

	// not a heartbeat request
	if len(param.Entries) > 0 {
		// 2. reply false if log dose’t contain an entry at prevLogIndex whose term matches prevLogTerm
		if param.PrevLogIndex >= 0 {
			if param.PrevLogIndex > p.GetLastLogIndex() || p.GetLog(param.PrevLogIndex).Term != param.PrevLogTerm {
				return proxy.Response{Term: p.GetTerm(), Success: false}
			}
		}
		p.AppendLog(param)
	}

	// 5. if leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if param.LeaderCommit > p.GetCommitIndex() {
		newIndex := param.LeaderCommit
		if p.GetLastLogIndex() < newIndex {
			newIndex = p.GetLastLogIndex()
		}
		p.SetCommitIndex(newIndex)
	}

	return proxy.Response{Term: p.GetTerm(), Success: true}
}

func (p *Follower) OnRequestVote(param proxy.RequestVote) proxy.Response {
	p.resetTimer()

	term := p.GetTerm()
	if param.Term < term {
		return proxy.Response{Term: term, Success: false}
	}

	// a new round of election
	if param.Term > term {
		p.SetTerm(param.Term)
		p.votedFor = ""
	}
	term = p.GetTerm()

	// have vote for other candidate in this term
	if len(p.votedFor) > 0 && p.votedFor != param.CandidateID {
		return proxy.Response{Term: p.GetTerm(), Success: false}
	}

	lastLogIndex := p.GetLastLogIndex()
	if lastLogIndex >= 0 {
		lastLogTerm := p.GetLog(lastLogIndex).Term
		if lastLogTerm > param.LastLogTerm {
			return proxy.Response{Term: p.GetTerm(), Success: false}
		}
		if lastLogTerm == param.LastLogTerm {
			if lastLogIndex > param.LastLogIndex {
				return proxy.Response{Term: term, Success: false}
			}
		}
	}

	p.votedFor = param.CandidateID
	return proxy.Response{Term: term, Success: true}
}

func (p *Follower) Timeout() {
	p.TransferState(StateEnum.Candidate)
}

// reset timer for follower, as time configured
func (p *Follower) resetTimer() {
	p.SetTimer(p.GetConfig().GetReplicateTimeout())
}

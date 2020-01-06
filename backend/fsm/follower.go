package fsm

import (
	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/proxy"
)

type Follower struct {
	Prober
	votedFor    string
	leaderID    string
	stateLogger *common.Logger
}

func NewFollower(s Prober, logger *common.Logger) *Follower {
	return &Follower{s, "", "", logger.With("state", StateEnum.Follower)}
}

func (p *Follower) implStateInterface() {
	var _ State = &Follower{}
}

func (p *Follower) GetLogger() *common.Logger {
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
	_ = p.SetTerm(param.Term)
	p.leaderID = param.LeaderID

	// not a heartbeat request
	if len(param.Entries) > 0 {
		// 2. reply false if log dose’t contain an entry at prevLogIndex whose term matches prevLogTerm
		if param.PrevLogIndex >= 0 {
			log, err := p.GetLog(param.PrevLogIndex)
			if err != nil {
				p.stateLogger.Error(err)
				return proxy.Response{Term: p.GetTerm(), Success: false}
			}
			if param.PrevLogIndex > p.GetLastLogIndex() || log.Term != param.PrevLogTerm {
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
		_ = p.SetCommitIndex(newIndex)
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
		_ = p.SetTerm(param.Term)
		p.votedFor = ""
	}
	term = p.GetTerm()

	// have vote for other candidate in this term
	if len(p.votedFor) > 0 && p.votedFor != param.CandidateID {
		return proxy.Response{Term: p.GetTerm(), Success: false}
	}

	lastLogIndex := p.GetLastLogIndex()
	if lastLogIndex >= 0 {
		log, err := p.GetLog(lastLogIndex)
		if err != nil {
			p.stateLogger.Error(err)
			return proxy.Response{Term: term, Success: false}
		}
		lastLogTerm := log.Term
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

func (p *Follower) GetLeader() string {
	return p.leaderID
}

func (p *Follower) GetVoteFor() string {
	return p.votedFor
}

// reset timer for follower, as time configured
func (p *Follower) resetTimer() {
	_ = p.SetTimer(p.GetConfig().GetReplicateTimeout())
}

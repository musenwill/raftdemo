package fsm

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
)

type Follower struct {
	*Server  // embed server
	leaderID string
}

func NewFollower(s *Server, conf *config.Config) *Follower {
	return &Follower{s, ""}
}

func (p *Follower) implStateInterface() {
	var _ State = &Follower{}
}

func (p *Follower) enterState() {
	p.resetTimer()
	p.votedFor = ""
}

func (p *Follower) leaveState() {
}

func (p *Follower) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	p.resetTimer()

	// 1. reply false if term < currentTerm
	if param.Term < p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	}
	p.currentTerm = param.Term
	p.leaderID = param.LeaderID

	// 2. reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
	if param.PrevLogIndex >= 0 {
		if param.PrevLogIndex > p.lastLogIndex() || p.logs[param.PrevLogIndex].Term != param.PrevLogTerm {
			return proxy.Response{Term: p.currentTerm, Success: false}
		}
	}

	if param.PrevLogIndex < p.getCommitIndex()-1 {
		panic("AppendEntries.PrevLogIndex < Server.commitIndex - 1, this is not allowed")
	}

	// 3. if an existing entry conflicts with a new one (same index but different terms),
	//    delete the existing entry and all that follow it
	p.logs = p.logs[:param.PrevLogIndex+1]

	// 4. Append any new entries not already in the log
	for _, g := range param.Entries {
		p.logs = append(p.logs, &g)
	}

	// 5. if leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if param.LeaderCommit > p.getCommitIndex() {
		newIndex := param.LeaderCommit
		if p.lastLogIndex() < newIndex {
			newIndex = p.lastLogIndex()
		}
		p.setCommitIndex(newIndex)
	}

	return proxy.Response{Term: p.currentTerm, Success: true}
}

func (p *Follower) onRequestVote(param proxy.RequestVote) proxy.Response {
	p.resetTimer()

	if param.Term < p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	}

	// a new round of election
	if param.Term > p.currentTerm {
		p.currentTerm = param.Term
		p.votedFor = ""
	}

	// have vote for other candidate in this term
	if len(p.votedFor) > 0 && p.votedFor != param.CandidateID {
		return proxy.Response{Term: p.currentTerm, Success: false}
	}

	if param.LastLogIndex <= p.lastLogIndex() {
		if param.LastLogIndex >= 0 && p.logs[param.LastLogIndex].Term != param.Term {
			return proxy.Response{Term: p.currentTerm, Success: false}
		}
	} else {
		return proxy.Response{Term: p.currentTerm, Success: false}
	}

	p.votedFor = param.CandidateID

	return proxy.Response{Term: p.currentTerm, Success: true}
}

func (p *Follower) timeout() {
	p.transferState(NewCandidate(p.Server, p.config))
}

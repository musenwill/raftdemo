package fsm

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type Follower struct {
	*Server     // embed server
	leaderID    string
	stateLogger *zap.SugaredLogger
}

func NewFollower(s *Server, conf *config.Config) *Follower {
	return &Follower{s, "", s.logger.With("state", "follower")}
}

func (p *Follower) implStateInterface() {
	var _ State = &Follower{}
}

func (p *Follower) enterState() {
	p.stateLogger.Infow("enter state", "term", p.currentTerm, "voteFor", p.votedFor,
		"commitIndex", p.commitIndex, "lastAplied", p.lastAplied)
	p.resetTimer()
	p.votedFor = ""
}

func (p *Follower) leaveState() {
	p.stateLogger.Infow("leave state", "term", p.currentTerm, "voteFor", p.votedFor,
		"commitIndex", p.commitIndex, "lastAplied", p.lastAplied)
}

func (p *Follower) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	p.resetTimer()

	// 1. reply false if term < currentTerm
	if param.Term < p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	}
	p.currentTerm = param.Term
	p.leaderID = param.LeaderID

	// not a heartbeat request
	if len(param.Entries) > 0 {
		// 2. reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
		if param.PrevLogIndex >= 0 {
			if param.PrevLogIndex > p.lastLogIndex() || p.logs[param.PrevLogIndex].Term != param.PrevLogTerm {
				return proxy.Response{Term: p.currentTerm, Success: false}
			}
		}

		// 3. if an existing entry conflicts with a new one (same index but different terms),
		//    delete the existing entry and all that follow it
		p.logs = p.logs[:param.PrevLogIndex+1]

		// 4. Append any new entries not already in the log
		p.logs = append(p.logs, param.Entries...)
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

	if p.lastLogIndex() >= 0 {
		if p.logs[p.lastLogIndex()].Term > param.LastLogTerm {
			return proxy.Response{Term: p.currentTerm, Success: false}
		}
		if p.logs[p.lastLogIndex()].Term == param.LastLogTerm {
			if p.lastLogIndex() > param.LastLogIndex {
				return proxy.Response{Term: p.currentTerm, Success: false}
			}
		}
	}

	p.votedFor = param.CandidateID
	return proxy.Response{Term: p.currentTerm, Success: true}
}

func (p *Follower) timeout() {
	p.transferState(NewCandidate(p.Server, p.config))
}

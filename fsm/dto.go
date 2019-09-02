package fsm

type Log struct {
	RequestID   string
	CandidateID string
	Term        int
}

type AppendEntries struct {
	Term         int
	LeaderID     string
	PrevLogIndex int
	prevLogTerm  int
	LeaderCommit int
	Entries      []Log
}

type RequestVote struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

type Response struct {
	Term    int
	Success bool
}

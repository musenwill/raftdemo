package fsm

type Log struct {
	RequestID   string
	CandidateID string
	Term        int64
}

type AppendEntries struct {
	Term         int64
	LeaderID     string
	PrevLogIndex int64
	prevLogTerm  int64
	LeaderCommit int64
	Entries      []Log
}

type RequestVote struct {
	Term         int64
	CandidateID  string
	LastLogIndex int64
	LastLogTerm  int64
}

type Response struct {
	Term    int64
	Success bool
}

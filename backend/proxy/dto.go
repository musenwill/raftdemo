package proxy

import "github.com/musenwill/raftdemo/model"

type AppendEntries struct {
	Term         int64
	LeaderID     string
	PrevLogIndex int64
	PrevLogTerm  int64
	LeaderCommit int64
	Entries      []model.Log
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

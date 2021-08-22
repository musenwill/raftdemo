package raft

import (
	"github.com/musenwill/raftdemo/model"
)

type Node struct {
	ID string
}

type State interface {
	Enter()
	Leave()
	OnAppendEntries(param model.AppendEntries) model.Response
	OnRequestVote(param model.RequestVote) model.Response
	OnTimeout()
}

type NodeInstance interface {
	Open() error
	Close() error

	GetNodeID() string
	GetTerm() int64
	GetCommitIndex() int64
	GetLastLogIndex() int64
	GetLastLogTerm() int64
	GetLastAppliedIndex() int64

	GetState() model.StateRole
	SwitchStateTo(state model.StateRole) error

	AppendData(data []byte) (model.Entry, error)
	AppendEntries(entries []model.AppendEntries) error
	GetEntries() []model.Entry
	GetEntry(index int64) (model.Entry, error)

	GetLeader() string
	GetVoteFor() string
}

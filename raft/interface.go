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
	State() model.StateRole
}

type NodeInstance interface {
	Open() error
	Close() error

	GetNodeID() string
	GetTerm() int64
	// CompareAndSetTerm set term is param is greater than current term
	// return -1 if param < current term
	// return 0 if param == current term
	// return 1 if param > current term
	CompareAndSetTerm(int64) int
	IncreaseTerm()
	GetCommitIndex() int64
	// CompareAndSetCommitIndex set commitID is param is greater than current commitID
	// return -1 if param < current CI
	// return 0 if param == current CI
	// return 1 if param > current CI
	CompareAndSetCommitIndex(int64) int
	GetLastLogIndex() int64
	GetLastLogTerm() int64
	GetLastAppliedIndex() int64

	GetState() model.StateRole
	SwitchStateTo(state model.StateRole) error

	AppendData(data []byte) error
	AppendEntries(entries []model.AppendEntries) error
	GetEntries() []model.Entry
	GetEntry(index int64) (model.Entry, error)

	GetLeader() string
	GetVoteFor() string
	SetLeader(string)

	OnAppendEntries(param model.AppendEntries) model.Response
	OnRequestVote(param model.RequestVote) model.Response
}

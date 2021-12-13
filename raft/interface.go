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
	// CASTerm set term is param is greater than current term
	// return -1 if param < current term
	// return 0 if param == current term
	// return 1 if param > current term
	CASTerm(term int64) int
	IncreaseTerm()
	GetCommitIndex() int64
	// CASCommitID set commitID is param is greater than current commitID
	// return -1 if param < current CI
	// return 0 if param == current CI
	// return 1 if param > current CI
	CASCommitID(int64) int
	GetLastLogIndex() int64
	GetLastLogTerm() int64
	GetLastAppliedIndex() int64

	GetState() model.StateRole
	SwitchStateTo(state model.StateRole) error

	AppendNop()
	AppendData(data []byte) error
	AppendEntries(entries []*model.Entry) error
	// GetFollowingEntries returns entries which >= index
	GetFollowingEntries(index int64) []*model.Entry
	GetEntry(index int64) (model.Entry, error)

	// WaitApply waits until term of the last applied entry is equal to the current term
	WaitApply(abort chan bool)

	GetLeader() string
	GetVoteFor() string
	SetLeader(string)
	SetVoteFor(string)
	RestLeader()
	RestVoteFor()

	SetReadable(readable bool)

	OnAppendEntries(param model.AppendEntries) model.Response
	OnRequestVote(param model.RequestVote) model.Response

	Broadcast(name string, abort chan bool, getRequest func(string) (interface{}, error), handleResponse func(string, model.Response))
}

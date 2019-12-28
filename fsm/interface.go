package fsm

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type StateName string

var StateEnum = struct {
	None, Follower, Candidate, Leader StateName
}{
	None:      "None",
	Follower:  "Follower",
	Candidate: "Candidate",
	Leader:    "Leader",
}

type State interface {
	EnterState()
	LeaveState()
	OnAppendEntries(param proxy.AppendEntries) proxy.Response
	OnRequestVote(param proxy.RequestVote) proxy.Response
	Timeout()
	Loggable

	GetLeader() string
	GetVoteFor() string
}

type Prober interface {
	Start()
	Stop()

	SetTimer(time int64) error // milliseconds
	ResetTimer()

	GetHost() string

	GetState() StateName
	GetCurrentState() State
	TransferState(state StateName)

	GetTerm() int64
	SetTerm(i int64) error
	IncreaseTerm()

	GetCommitIndex() int64
	SetCommitIndex(i int64) error

	GetLastAppliedIndex() int64
	SetLastAppliedIndex(i int64) error
	IncreaseLastAppliedIndex()

	GetLastLogIndex() int64
	GetLogs() []model.Log
	GetLog(index int64) (model.Log, error)
	AppendLog(entries proxy.AppendEntries)
	AddLogs(logs ...model.Log) ([]model.Log, error)

	GetLeader() string
	GetVoteFor() string

	GetConfig() config.Config
	GetProxy() proxy.Proxy
}

type Loggable interface {
	GetLogger() *zap.SugaredLogger
}

package fsm

import (
	"github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
)

type Instance struct {
	nodeID string
	State  model.StateRole
}

func NewInstance(nodeID string, committer committer.Committer, proxy proxy.Proxy, cfg *raft.Config) *Instance {
	return &Instance{nodeID: nodeID}
}

func (s *Instance) Open() error {
	return nil
}

func (s *Instance) Close() error {
	return nil
}

func (s *Instance) GetNodeID() string {
	return s.nodeID
}

func (s *Instance) GetTerm() int64 {
	return 0
}

func (s *Instance) GetCommitIndex() int64 {
	return 0
}

func (s *Instance) GetLastLogIndex() int64 {
	return 0
}

func (s *Instance) GetLastLogTerm() int64 {
	return 0
}

func (s *Instance) GetLastAppliedIndex() int64 {
	return 0
}

func (s *Instance) GetState() model.StateRole {
	return s.State
}

func (s *Instance) SwitchStateTo(state model.StateRole) error {
	return nil
}

func (s *Instance) AppendData(data []byte) (model.Entry, error) {
	return model.Entry{}, nil
}

func (s *Instance) AppendEntries(entries []model.AppendEntries) error {
	return nil
}

func (s *Instance) GetEntries() []model.Entry {
	return make([]model.Entry, 0)
}

func (s *Instance) GetEntry(index int64) (model.Entry, error) {
	return model.Entry{}, nil
}

func (s *Instance) GetLeader() string {
	return ""
}

func (s *Instance) GetVoteFor() string {
	return ""
}

package fsm

import (
	"fmt"
	"sync"

	"github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
)

const MaxRequests = 1000

type Instance struct {
	nodeID string
	nodes  []string

	term      int64
	commitID  int64
	appliedID int64
	entries   []model.Entry

	state     raft.State
	proxy     proxy.Proxy
	committer committer.Committer
	cfg       *raft.Config

	requestWaiters map[int64]chan error

	mu sync.RWMutex
}

func NewInstance(nodeID string, nodes []string, committer committer.Committer, proxy proxy.Proxy, cfg *raft.Config) *Instance {
	var nodeIDs []string
	for _, n := range nodes {
		if n != nodeID {
			nodeIDs = append(nodeIDs, n)
		}
	}

	return &Instance{
		nodeID:    nodeID,
		nodes:     nodeIDs,
		entries:   make([]model.Entry, 0),
		proxy:     proxy,
		committer: committer,
		cfg:       cfg,

		requestWaiters: make(map[int64]chan error),
	}
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.term
}

func (s *Instance) SetTerm(term int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if term > s.term {
		s.term = term
	}
}

func (s *Instance) IncreaseTerm() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.term++
}

func (s *Instance) GetCommitIndex() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.commitID
}

func (s *Instance) SetCommitIndex(commitID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastLog := s.getLastLog()
	if commitID > lastLog.Id {
		commitID = lastLog.Id
	}

	if commitID > s.commitID {
		s.commitID = commitID
	}
}

func (s *Instance) GetLastLogIndex() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getLastLog().Id
}

func (s *Instance) GetLastLogTerm() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getLastLog().Term
}

func (s *Instance) GetLastAppliedIndex() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.appliedID
}

func (s *Instance) GetState() model.StateRole {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state.State()
}

func (s *Instance) SwitchStateTo(state model.StateRole) error {
	return nil
}

func (s *Instance) AppendData(data []byte) error {
	errChan := make(chan error)
	defer func() {
		close(errChan)
	}()

	if err := func() error {
		s.mu.Lock()
		defer s.mu.Unlock()

		if len(s.requestWaiters) >= MaxRequests {
			return fmt.Errorf("pending requests exceeded")
		}

		entry, err := s.appendData(data)
		if err != nil {
			return err
		}

		s.requestWaiters[entry.Id] = errChan
		return nil
	}(); err != nil {
		return err
	}

	return <-errChan
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

func (s *Instance) getLastLog() model.Entry {
	if len(s.entries) == 0 {
		return model.Entry{}
	}

	return s.entries[len(s.entries)-1]
}

func (s *Instance) appendData(data []byte) (model.Entry, error) {
	lastLog := s.getLastLog()

	entry := model.Entry{
		Id:      lastLog.Id + 1,
		Term:    s.term,
		Type:    model.EntryType_Data,
		Payload: data,
	}

	s.entries = append(s.entries, entry)

	return entry, nil
}

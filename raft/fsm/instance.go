package fsm

import (
	"fmt"
	"sync"

	"github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/log"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const MaxRequests = 1000

var NodeClosedErr = fmt.Errorf("node instance closed")

type Instance struct {
	nodeID string
	nodes  []string

	term      atomic.Int64
	commitID  int64
	appliedID atomic.Int64
	entries   []model.Entry
	leader    string
	voteFor   string

	state     raft.State
	proxy     proxy.Proxy
	committer committer.Committer
	cfg       *raft.Config

	requestWaiters map[int64]chan error

	mu               sync.RWMutex
	wg               sync.WaitGroup
	commitIDUpdateCh chan int64
	closing          chan bool

	logger log.Logger
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

		requestWaiters:   make(map[int64]chan error),
		commitIDUpdateCh: make(chan int64),
		closing:          make(chan bool),
	}
}

func (s *Instance) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger = *s.cfg.Logger.With(zap.String("nodeID", s.nodeID))
	s.entries = append(s.entries, model.Entry{}) // empty entry in index 0
	s.receiveJob()
	s.commitJob()

	return nil
}

func (s *Instance) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

	close(s.closing)
	close(s.commitIDUpdateCh)
	s.state.Leave()
	s.wg.Wait()

	return nil
}

func (s *Instance) receiveJob() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.closing:
				return
			default:
			}

			if err := s.proxy.Receive(s.nodeID, s.handleRequest); err != nil {
				s.logger.Fatalf("receive from proxy %w", err)
			}
		}
	}()
}

func (s *Instance) commitJob() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.closing:
				return
			case CI := <-s.commitIDUpdateCh:
				for appliedID := s.appliedID.Load(); appliedID < CI; s.appliedID.Inc() {
					err := s.committer.Commit(s.entries[appliedID+1].Payload)

					if errCh, ok := s.requestWaiters[appliedID+1]; ok {
						errCh <- err
						delete(s.requestWaiters, appliedID+1)
					}

					if err != nil {
						break
					}
				}
			}
		}
	}()
}

func (s *Instance) handleRequest(request interface{}) model.Response {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	switch t := request.(type) {
	case model.AppendEntries:
		return state.OnAppendEntries(t)
	case model.RequestVote:
		return state.OnRequestVote(t)
	default:
		s.logger.Fatalf("unknown request type %t", t)
		return model.Response{}
	}
}

func (s *Instance) GetNodeID() string {
	return s.nodeID
}

func (s *Instance) GetTerm() int64 {
	return s.term.Load()
}

func (s *Instance) SetTerm(term int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if term > s.term.Load() {
		s.term.Store(term)
		if err := s.SwitchStateTo(model.StateRole_Follower); err != nil {
			s.logger.Fatalf("switch state to follower on larger term, %w", err)
		}
	}
}

func (s *Instance) IncreaseTerm() {
	s.term.Inc()
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
		s.commitIDUpdateCh <- commitID
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
	return s.appliedID.Load()
}

func (s *Instance) GetState() model.StateRole {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state.State()
}

func (s *Instance) SwitchStateTo(state model.StateRole) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

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

		select {
		case <-s.closing:
			return NodeClosedErr
		default:
		}

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
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

	return nil
}

func (s *Instance) GetEntries() []model.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.entries
}

func (s *Instance) GetEntry(index int64) (model.Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index >= int64(len(s.entries)) {
		return model.Entry{}, fmt.Errorf("get entry index out of range")
	}

	return s.entries[index], nil
}

func (s *Instance) GetLeader() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.leader
}

func (s *Instance) SetLeader(leader string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.leader = leader
}

func (s *Instance) GetVoteFor() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.voteFor
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
		Term:    s.term.Load(),
		Type:    model.EntryType_Data,
		Payload: data,
	}

	s.entries = append(s.entries, entry)

	return entry, nil
}

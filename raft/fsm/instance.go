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
	commitID  atomic.Int64
	appliedID atomic.Int64
	entries   []model.Entry
	leader    string
	voteFor   string

	state     raft.State
	proxy     proxy.Proxy
	committer committer.Committer
	cfg       *raft.Config

	requestWaiters map[int64]chan error

	mu                sync.RWMutex
	wg                sync.WaitGroup
	commitIDUpdateCh  chan int64
	appliedCIUpdateCh chan int64
	closing           chan bool
	readable          atomic.Bool

	logger log.Logger
}

var _ raft.NodeInstance = &Instance{}

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

		requestWaiters:    make(map[int64]chan error),
		commitIDUpdateCh:  make(chan int64),
		appliedCIUpdateCh: make(chan int64),
		closing:           make(chan bool),
	}
}

func (s *Instance) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger = *s.cfg.Logger.With(zap.String("nodeID", s.nodeID))
	s.entries = append(s.entries, model.Entry{}) // empty entry in index 0

	s.state = NewFollower(s, s.cfg, s.logger)
	s.state.Enter()

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

			if err := s.proxy.Receive(s.nodeID, s.handleRequest, s.closing); err != nil {
				s.printLog(s.logger.Error, "receive from proxy", err)
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
				for appliedID := s.appliedID.Load(); appliedID < CI; {
					entry := s.entries[appliedID+1]
					if entry.Type == model.EntryType_Data {
						err := s.committer.Commit(entry.Payload)
						if errCh, ok := s.requestWaiters[appliedID+1]; ok {
							errCh <- err
							delete(s.requestWaiters, appliedID+1)
						}

						if err != nil {
							s.printLog(s.logger.Error, fmt.Sprintf("commit log index %d", entry.Id), err)
							break
						}
					}
					s.appliedID.Inc()
					select {
					case s.appliedCIUpdateCh <- s.appliedID.Load():
					default:
					}
				}
			}
		}
	}()
}

func (s *Instance) handleRequest(request interface{}) model.Response {
	switch t := request.(type) {
	case model.AppendEntries:
		return s.OnAppendEntries(t)
	case model.RequestVote:
		return s.OnRequestVote(t)
	default:
		s.printLog(s.logger.Fatal, fmt.Sprintf("unknown request type %t", t), nil)
		return model.Response{}
	}
}

func (s *Instance) OnAppendEntries(request model.AppendEntries) model.Response {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	return state.OnAppendEntries(request)
}

func (s *Instance) OnRequestVote(request model.RequestVote) model.Response {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	return state.OnRequestVote(request)
}

func (s *Instance) GetNodeID() string {
	return s.nodeID
}

func (s *Instance) GetTerm() int64 {
	return s.term.Load()
}

func (s *Instance) CASTerm(term int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if term > s.term.Load() {
		s.term.Store(term)
		s.voteFor = ""
		return 1
	} else if term == s.term.Load() {
		return 0
	} else {
		return -1
	}
}

func (s *Instance) IncreaseTerm() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.term.Inc()
	s.voteFor = ""
}

func (s *Instance) GetCommitIndex() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.commitID.Load()
}

func (s *Instance) CASCommitID(commitID int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastLog := s.getLastLog()
	if commitID > lastLog.Id {
		commitID = lastLog.Id
	}

	if commitID > s.commitID.Load() {
		s.commitID.Store(commitID)
		s.commitIDUpdateCh <- commitID
		return 1
	} else if commitID == s.commitID.Load() {
		return 0
	} else {
		return -1
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
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

	return nil
}

func (s *Instance) AppendNop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastLog := s.getLastLog()

	entry := model.Entry{
		Id:      lastLog.Id + 1,
		Term:    s.term.Load(),
		Type:    model.EntryType_Nop,
		Payload: nil,
	}

	s.entries = append(s.entries, entry)
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

func (s *Instance) AppendEntries(entries []*model.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

	index := len(s.entries) - 1
	for index >= 0 {
		if s.entries[index].Term > entries[0].Term {
			index--
			continue
		}
		if s.entries[index].Id >= entries[0].Id {
			index--
			continue
		}
	}
	s.entries = s.entries[:index+1]

	for _, e := range entries {
		s.entries = append(s.entries, *e)
	}

	return nil
}

func (s *Instance) GetFollowingEntries(index int64) []*model.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]*model.Entry, 0)

	if index <= 1 {
		index = 1
	} else if index >= int64(len(s.entries)) {
		return entries
	}

	following := s.entries[index:]
	for _, e := range following {
		entries = append(entries, &e)
	}

	return entries
}

func (s *Instance) GetEntry(index int64) (model.Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index >= int64(len(s.entries)) {
		return model.Entry{}, fmt.Errorf("get entry index out of range")
	}

	return s.entries[index], nil
}

func (s *Instance) WaitApply(abort chan bool) {
	for {
		select {
		case <-abort:
			return
		case applied := <-s.appliedCIUpdateCh:
			s.mu.RLock()
			entry := s.entries[applied]
			s.mu.RUnlock()

			if entry.Term == s.term.Load() {
				return
			}
		}
	}
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

func (s *Instance) SetVoteFor(voteFor string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.voteFor = voteFor
}

func (s *Instance) RestLeader() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.leader = ""
}

func (s *Instance) RestVoteFor() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.voteFor = ""
}

func (s *Instance) SetReadable(readable bool) {
	s.readable.Store(readable)
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

func (s *Instance) printLog(fn func(args ...interface{}), msg string, err error) {
	fn(zap.Int64("term", s.term.Load()),
		zap.Int64("commitID", s.commitID.Load()),
		zap.Int64("appliedID", s.appliedID.Load()),
		zap.String("state", s.state.State().String()),
		zap.Bool("readable", s.readable.Load()),
		zap.String("leader", s.leader),
		zap.String("vote for", s.voteFor),
		zap.String("msg", msg), zap.Error(err))
}

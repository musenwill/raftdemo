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

	mu struct {
		sync.RWMutex
		term     atomic.Int64
		commitID atomic.Int64
		lastID   atomic.Int64
		entries  []model.Entry
		voteFor  string
	}

	appliedID atomic.Int64

	muLeader struct {
		sync.RWMutex
		leader string
	}

	muState struct {
		sync.RWMutex
		state raft.State
	}

	muRequests struct {
		sync.Mutex
		requestWaiters map[int64]chan error
	}

	proxy     proxy.Proxy
	committer committer.Committer
	cfg       *raft.Config

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

	in := &Instance{
		nodeID:    nodeID,
		nodes:     nodeIDs,
		proxy:     proxy,
		committer: committer,
		cfg:       cfg,

		commitIDUpdateCh:  make(chan int64, 1),
		appliedCIUpdateCh: make(chan int64, 1),
		closing:           make(chan bool),
	}

	in.mu.entries = make([]model.Entry, 0)
	in.muRequests.requestWaiters = make(map[int64]chan error)

	return in
}

func (s *Instance) Open() error {
	s.logger = *s.cfg.Logger.With(zap.String("nodeID", s.nodeID))
	s.mu.entries = append(s.mu.entries, model.Entry{}) // empty entry in index 0
	s.muState.state = NewFollower(s, s.cfg, s.logger)

	s.muState.state.Enter()
	s.receiveJob()
	s.commitJob()

	return nil
}

func (s *Instance) Close() error {
	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

	close(s.closing)
	close(s.commitIDUpdateCh)
	s.muState.state.Leave()
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
					// the entries before ci is readonly, can't have conflict access, so no need to wrap in lock
					entry := s.mu.entries[appliedID+1]
					if entry.Type == model.EntryType_Data {
						err := s.committer.Commit(entry.Payload)
						if errCh, ok := s.muRequests.requestWaiters[appliedID+1]; ok {
							errCh <- err
							s.muRequests.Lock()
							delete(s.muRequests.requestWaiters, appliedID+1)
							s.muRequests.Unlock()
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
	s.muState.RLock()
	state := s.muState.state
	s.muState.RUnlock()

	return state.OnAppendEntries(request)
}

func (s *Instance) OnRequestVote(request model.RequestVote) model.Response {
	s.muState.RLock()
	state := s.muState.state
	s.muState.RUnlock()

	return state.OnRequestVote(request)
}

func (s *Instance) GetNodeID() string {
	return s.nodeID
}

func (s *Instance) GetTerm() int64 {
	return s.mu.term.Load()
}

func (s *Instance) CASTerm(term int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if term > s.mu.term.Load() {
		s.mu.term.Store(term)
		s.mu.voteFor = ""
		return 1
	} else if term == s.mu.term.Load() {
		return 0
	} else {
		return -1
	}
}

func (s *Instance) IncreaseTerm() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.term.Inc()
	s.mu.voteFor = ""
}

func (s *Instance) GetCommitIndex() int64 {
	return s.mu.commitID.Load()
}

func (s *Instance) CASCommitID(commitID int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastLog := s.mu.lastID.Load()
	if commitID > lastLog {
		commitID = lastLog
	}

	if commitID > s.mu.commitID.Load() {
		s.mu.commitID.Store(commitID)
		s.commitIDUpdateCh <- commitID
		return 1
	} else if commitID == s.mu.commitID.Load() {
		return 0
	} else {
		return -1
	}
}

func (s *Instance) GetLastAppliedIndex() int64 {
	return s.appliedID.Load()
}

func (s *Instance) GetState() model.StateRole {
	s.muState.RLock()
	defer s.muState.RUnlock()

	return s.muState.state.State()
}

func (s *Instance) SwitchStateTo(state model.StateRole) error {
	s.muState.Lock()
	defer s.muState.Unlock()

	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

	return nil
}

func (s *Instance) AppendNop() {
	entry := model.Entry{
		Id:      s.mu.lastID.Load() + 1,
		Term:    s.mu.term.Load(),
		Type:    model.EntryType_Nop,
		Payload: nil,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.entries = append(s.mu.entries, entry)
	s.mu.lastID.Inc()
}

func (s *Instance) AppendData(data []byte) error {
	errChan := make(chan error)
	defer func() {
		close(errChan)
	}()

	if err := func() error {
		select {
		case <-s.closing:
			return NodeClosedErr
		default:
		}

		s.muRequests.Lock()
		defer s.muRequests.Unlock()

		if len(s.muRequests.requestWaiters) >= MaxRequests {
			return fmt.Errorf("pending requests exceeded")
		}

		s.mu.Lock()
		entry, err := s.appendData(data)
		if err != nil {
			s.mu.Unlock()
			return err
		}
		s.mu.Unlock()

		s.muRequests.requestWaiters[entry.Id] = errChan
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

	select {
	case <-s.closing:
		return NodeClosedErr
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	index := len(s.mu.entries) - 1
	for index >= 0 {
		if s.mu.entries[index].Term > entries[0].Term {
			index--
			continue
		}
		if s.mu.entries[index].Id >= entries[0].Id {
			index--
			continue
		}
	}
	s.mu.entries = s.mu.entries[:index+1]

	for _, e := range entries {
		s.mu.entries = append(s.mu.entries, *e)
	}

	s.mu.lastID.Store(s.mu.entries[len(s.mu.entries)-1].Id)

	return nil
}

func (s *Instance) GetFollowingEntries(index int64) []*model.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]*model.Entry, 0)

	if index <= 1 {
		index = 1
	} else if index >= int64(len(s.mu.entries)) {
		return entries
	}

	following := s.mu.entries[index:]
	for _, e := range following {
		entries = append(entries, &e)
	}

	return entries
}

func (s *Instance) GetEntry(index int64) (model.Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index >= int64(len(s.mu.entries)) {
		return model.Entry{}, fmt.Errorf("get entry index out of range")
	}

	return s.mu.entries[index], nil
}

func (s *Instance) GetLastEntry() model.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getLastEntry()
}

func (s *Instance) WaitApply(abort chan bool) {
	for {
		select {
		case <-abort:
			return
		case applied := <-s.appliedCIUpdateCh:
			s.mu.RLock()
			entry := s.mu.entries[applied]
			s.mu.RUnlock()

			if entry.Term == s.mu.term.Load() {
				return
			}
		}
	}
}

func (s *Instance) GetLeader() string {
	s.muLeader.RLock()
	defer s.muLeader.RUnlock()

	return s.muLeader.leader
}

func (s *Instance) SetLeader(leader string) {
	s.muLeader.Lock()
	defer s.muLeader.Unlock()

	s.muLeader.leader = leader
}

func (s *Instance) GetVoteFor() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.mu.voteFor
}

func (s *Instance) SetVoteFor(voteFor string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.voteFor = voteFor
}

func (s *Instance) RestLeader() {
	s.muLeader.Lock()
	defer s.muLeader.Unlock()

	s.muLeader.leader = ""
}

func (s *Instance) RestVoteFor() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.voteFor = ""
}

func (s *Instance) SetReadable(readable bool) {
	s.readable.Store(readable)
}

func (s *Instance) Broadcast(name string, abort chan bool, getRequest func(string) (interface{}, error), handleResponse func(string, model.Response)) {
	go func() {
		for _, n := range s.nodes {
			nodeID := n
			select {
			case <-abort:
				return
			default:
			}
			go func() {
				request, err := getRequest(nodeID)
				if err != nil {
					s.logger.Errorf("%s, %w", name, err)
					return
				}
				select {
				case <-abort:
					return
				default:
				}
				response, err := s.proxy.Send(nodeID, request, abort)
				if err != nil {
					s.logger.Errorf("%s, %w", name, err)
					return
				}
				select {
				case <-abort:
					return
				default:
				}
				handleResponse(nodeID, response)
			}()
		}
	}()
}

func (s *Instance) getLastEntry() model.Entry {
	if len(s.mu.entries) == 0 {
		return model.Entry{}
	}

	return s.mu.entries[len(s.mu.entries)-1]
}

func (s *Instance) appendData(data []byte) (model.Entry, error) {
	entry := model.Entry{
		Id:      s.mu.lastID.Load() + 1,
		Term:    s.mu.term.Load(),
		Type:    model.EntryType_Data,
		Payload: data,
	}

	s.mu.entries = append(s.mu.entries, entry)

	return entry, nil
}

func (s *Instance) printLog(fn func(args ...interface{}), msg string, err error) {
	fn(zap.Int64("term", s.mu.term.Load()),
		zap.Int64("commitID", s.mu.commitID.Load()),
		zap.Int64("appliedID", s.appliedID.Load()),
		zap.String("state", s.muState.state.State().String()),
		zap.Bool("readable", s.readable.Load()),
		zap.String("leader", s.muLeader.leader),
		zap.String("vote for", s.mu.voteFor),
		zap.String("msg", msg), zap.Error(err))
}

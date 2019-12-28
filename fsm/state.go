package fsm

import (
	"errors"
	"fmt"
	"github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/model"
	"sync"
	"sync/atomic"
	"time"

	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type Server struct {
	id           string
	currentTerm  int64
	commitIndex  int64
	lastApplied  int64
	timer        *time.Timer
	logs         []model.Log
	currentState State
	state        StateName

	commitNotifier chan bool // notify to commit
	stopNotifier   chan bool // notify to stop the server

	config    config.Config
	logger    *zap.SugaredLogger
	committer committer.Committer
	proxy     proxy.Proxy

	logsLock  *sync.RWMutex
	stateLock *sync.RWMutex
}

func NewServer(id string, committer committer.Committer, proxy proxy.Proxy,
	conf config.Config, logger *zap.SugaredLogger) *Server {
	return &Server{
		id:          id,
		currentTerm: 0,
		commitIndex: -1, // in raft paper, an valid index begin with 1 in no empty logs. but in program, valid index usually begin with 0
		lastApplied: -1, // similar to commitIndex
		logs:        make([]model.Log, 0),

		commitNotifier: make(chan bool),
		stopNotifier:   make(chan bool),

		config:    conf,
		logger:    logger.With("node", id),
		committer: committer,
		proxy:     proxy,

		logsLock:  &sync.RWMutex{},
		stateLock: &sync.RWMutex{},
	}
}

func (p *Server) implLoggable() {
	var _ Loggable = &Server{}
}

func (p *Server) implProber() {
	var _ Prober = &Server{}
}

func (p *Server) Start() {
	p.TransferState(StateEnum.Follower)
	p.logger.Info("init state as follower")

	go p.commitTask()
	p.logger.Info("start up commit task")

	go p.fsmTask()
	p.logger.Info("start up fsm task")
}

func (p *Server) Stop() {
	p.TransferState(StateEnum.None)
	p.timer.Stop()

	close(p.stopNotifier)
	p.logger.Info("stop fsm task")

	close(p.commitNotifier)
	p.logger.Info("stop commit task")
}

func (p *Server) SetTimer(t int64) error {
	p.logger.Debugw(fmt.Sprintf("reset timer with %d milliseconds", t), "state", p.state, "term", p.currentTerm)

	if t <= 0 {
		return errors.New("timer should greater than 0")
	}

	if p.timer == nil {
		p.timer = time.NewTimer(time.Duration(int64(time.Millisecond) * t))
	} else {
		p.timer.Reset(time.Duration(int64(time.Millisecond) * t))
	}

	return nil
}

func (p *Server) ResetTimer() {
	p.GetCurrentState().Timeout()
}

func (p *Server) GetHost() string {
	return p.id
}

func (p *Server) GetState() StateName {
	p.stateLock.RLock()
	defer p.stateLock.RUnlock()

	return p.state
}

func (p *Server) GetCurrentState() State {
	p.stateLock.RLock()
	defer p.stateLock.RUnlock()

	return p.currentState
}

func (p *Server) GetTerm() int64 {
	return atomic.LoadInt64(&p.currentTerm)
}

func (p *Server) SetTerm(i int64) error {
	if i < p.GetTerm() {
		return errors.New("should not set term less than current term")
	}

	atomic.StoreInt64(&p.currentTerm, i)
	return nil
}

func (p *Server) IncreaseTerm() {
	atomic.AddInt64(&p.currentTerm, 1)
}

func (p *Server) GetCommitIndex() int64 {
	return atomic.LoadInt64(&p.commitIndex)
}

func (p *Server) SetCommitIndex(i int64) error {
	if i < p.GetCommitIndex() {
		return errors.New("should not set commit index less than current commit index")
	}

	atomic.StoreInt64(&p.commitIndex, i)
	p.commitNotifier <- true
	return nil
}

func (p *Server) GetLastAppliedIndex() int64 {
	return atomic.LoadInt64(&p.lastApplied)
}

func (p *Server) SetLastAppliedIndex(i int64) error {
	if i < p.GetLastAppliedIndex() {
		return errors.New("should not set applied index less than current applied index")
	}

	atomic.StoreInt64(&p.lastApplied, i)
	return nil
}

func (p *Server) IncreaseLastAppliedIndex() {
	atomic.StoreInt64(&p.lastApplied, 1)
}

func (p *Server) GetLastLogIndex() int64 {
	p.logsLock.RLock()
	defer p.logsLock.RUnlock()

	return int64(len(p.logs) - 1)
}

func (p *Server) GetLogs() []model.Log {
	p.logsLock.RLock()
	defer p.logsLock.RUnlock()

	return p.logs[:]
}

func (p *Server) GetLog(index int64) (model.Log, error) {
	p.logsLock.RLock()
	defer p.logsLock.RUnlock()

	if index < 0 || index >= int64(len(p.logs)) {
		return model.Log{}, errors.New("index out of bound")
	}

	return p.logs[index], nil
}

func (p *Server) AppendLog(entries proxy.AppendEntries) {
	// if an existing entry conflicts with a new one (same index but different terms),
	// delete the existing entry and all that follow it
	p.logs = p.logs[:entries.PrevLogIndex+1]

	// 4. Append any new entries not already in the log
	p.logs = append(p.logs, entries.Entries...)
}

func (p *Server) AddLogs(logs ...model.Log) ([]model.Log, error) {
	if p.GetState() != StateEnum.Leader {
		return nil, errors.New("can not add logs to none leader host")
	}

	term := p.GetTerm()
	var items []model.Log
	for _, l := range logs {
		items = append(items, model.Log{
			RequestID: l.RequestID,
			Command:   l.Command,
			Term:      term,
		})
	}

	p.logsLock.Lock()
	defer p.logsLock.Unlock()

	p.logs = append(p.logs, items...)
	return items, nil
}

func (p *Server) GetConfig() config.Config {
	return p.config
}

func (p *Server) GetProxy() proxy.Proxy {
	return p.proxy
}

func (p *Server) GetLeader() string {
	return p.GetCurrentState().GetLeader()
}

func (p *Server) GetVoteFor() string {
	return p.GetCurrentState().GetVoteFor()
}

func (p *Server) fsmTask() {
	for {
		appendEntriesRequestReader, err := p.proxy.AppendEntriesRequestReader(p.id)
		if err != nil {
			p.logger.Fatal(fmt.Errorf("%w", err))
			return
		}
		appendEntriesResponseSender, err := p.proxy.AppendEntriesResponseSender(p.id)
		if err != nil {
			p.logger.Fatal(fmt.Errorf("%w", err))
			return
		}
		voteRequestReader, err := p.proxy.VoteRequestReader(p.id)
		if err != nil {
			p.logger.Fatal(fmt.Errorf("%w", err))
			return
		}
		voteResponseSender, err := p.proxy.VoteResponseSender(p.id)
		if err != nil {
			p.logger.Fatal(fmt.Errorf("%w", err))
			return
		}

		select {
		case <-p.stopNotifier:
			return
		case _, ok := <-p.timer.C:
			if !ok {
				p.logger.Info("timer closed")
				return
			}
			p.logState()
			p.GetCurrentState().Timeout()
		case request, ok := <-appendEntriesRequestReader:
			if !ok {
				p.logger.Error(fmt.Sprintf("append entries request channel of node %s closed", p.id))
				break
			}

			p.logger.Debugw("receive append entries request", "state", p.state, "term", p.currentTerm,
				"commitIndex", p.commitIndex, "lastApplied", p.lastApplied, "body", request)
			response := p.GetCurrentState().OnAppendEntries(request)

			select {
			case <-p.stopNotifier:
				return
			case _, ok := <-p.timer.C:
				if !ok {
					p.logger.Info("timer closed")
					return
				}

				p.logState()
				p.GetCurrentState().Timeout()
			case appendEntriesResponseSender <- response:
				p.logger.Debugw("send append entries response", "state", p.state, "body", response)
			}
		case request, ok := <-voteRequestReader:
			if !ok {
				p.logger.Error(fmt.Sprintf("vote request channel of node %s closed", p.id))
				break
			}

			p.logger.Debugw("receive request vote request", "state", p.state, "term", p.currentTerm,
				"commitIndex", p.commitIndex, "lastApplied", p.lastApplied, "body", request)
			response := p.GetCurrentState().OnRequestVote(request)

			select {
			case <-p.stopNotifier:
				return
			case _, ok := <-p.timer.C:
				if !ok {
					p.logger.Info("timer closed")
					return
				}

				p.logState()
				p.GetCurrentState().Timeout()
			case voteResponseSender <- response:
				p.logger.Debugw("send vote response", "state", p.state, "body", response)
			}
		}
	}
}

func (p *Server) commitTask() {
	for {
		<-p.commitNotifier
		for i := p.lastApplied + 1; i <= p.GetCommitIndex(); i++ {
			err := p.committer.Commit(p.logs[i])
			if err != nil {
				p.logger.Errorw("commit log error", "state", p.state, "logIndex", i, "log", p.logs[i], "err", err)
				break
			}
			p.lastApplied++
			p.logger.Infow("succeed commit log", "state", p.state, "logIndex", i, "term", p.currentTerm,
				"commitIndex", p.commitIndex, "lastApplied", p.lastApplied)
		}
	}
}

func (p *Server) TransferState(state StateName) {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()

	if p.state == state {
		return
	}

	if p.currentState != nil {
		p.currentState.LeaveState()
	}

	p.state = state
	var newState State
	if state == StateEnum.Follower {
		newState = NewFollower(p, p.logger)
	} else if state == StateEnum.Candidate {
		newState = NewCandidate(p, p.logger)
	} else if state == StateEnum.Leader {
		newState = NewLeader(p, p.logger, p.config.GetNodes())
	} else {
		newState = NewDummyState(p)
	}
	p.currentState = newState
	p.currentState.EnterState()
}

func (p *Server) GetLogger() *zap.SugaredLogger {
	return p.logger
}

func (p *Server) logState() {
	p.logger.Infow("Timeout", "state", p.state, "term", p.currentTerm,
		"commitIndex", p.commitIndex, "lastApplied", p.lastApplied)
}

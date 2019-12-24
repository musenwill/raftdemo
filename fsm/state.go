package fsm

import (
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

	commitNotifier        chan bool      // notify to commit
	stopNotifier          chan bool      // notify to stop the server
	transferStateNotifier chan StateName // notify to transfer state

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

		commitNotifier:        make(chan bool),
		stopNotifier:          make(chan bool),
		transferStateNotifier: make(chan StateName),

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
	p.transferState(StateEnum.Follower)
	p.logger.Info("init state as follower")

	go p.commitTask()
	p.logger.Info("start up commit task")

	go p.fsmTask()
	p.logger.Info("start up fsm task")
}

func (p *Server) Stop() {
	p.transferState(StateEnum.None)
	p.timer.Stop()

	close(p.stopNotifier)
	p.logger.Info("stop fsm task")

	close(p.commitNotifier)
	p.logger.Info("stop commit task")
}

func (p *Server) SetTimer(t int64) {
	if p.timer == nil {
		p.timer = time.NewTimer(time.Duration(t))
	} else {
		p.timer.Reset(time.Duration(t))
	}
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

func (p *Server) NotifyTransferState(state StateName) {
	p.transferStateNotifier <- state
}

func (p *Server) GetTerm() int64 {
	return atomic.LoadInt64(&p.currentTerm)
}

func (p *Server) SetTerm(i int64) {
	atomic.StoreInt64(&p.currentTerm, i)
}

func (p *Server) IncreaseTerm() {
	atomic.AddInt64(&p.currentTerm, 1)
}

func (p *Server) GetCommitIndex() int64 {
	return atomic.LoadInt64(&p.commitIndex)
}

func (p *Server) SetCommitIndex(i int64) {
	atomic.StoreInt64(&p.commitIndex, i)
	p.commitNotifier <- true
}

func (p *Server) GetLastAppliedIndex() int64 {
	return atomic.LoadInt64(&p.lastApplied)
}

func (p *Server) SetLastAppliedIndex(i int64) {
	atomic.StoreInt64(&p.lastApplied, i)
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

func (p *Server) GetLog(index int64) model.Log {
	p.logsLock.RLock()
	defer p.logsLock.RUnlock()

	return p.logs[index]
}

func (p *Server) AppendLog(entries proxy.AppendEntries) {
	// if an existing entry conflicts with a new one (same index but different terms),
	// delete the existing entry and all that follow it
	p.logs = p.logs[:entries.PrevLogIndex+1]

	// 4. Append any new entries not already in the log
	p.logs = append(p.logs, entries.Entries...)
}

func (p *Server) GetConfig() config.Config {
	return p.config
}

func (p *Server) GetProxy() proxy.Proxy {
	return p.proxy
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
			p.currentState.Timeout()
		case stateName, ok := <-p.transferStateNotifier:
			if !ok {
				p.logger.Info("state transfer notifier closed")
				return
			}
			p.transferState(stateName)
		case request, ok := <-appendEntriesRequestReader:
			if !ok {
				p.logger.Error(fmt.Sprintf("append entries request channel of node %s closed", p.id))
				break
			}

			p.currentState.GetLogger().Debugw("receive append entries request", "term", p.currentTerm,
				"commitIndex", p.commitIndex, "lastApplied", p.lastApplied, "body", request)
			response := p.currentState.OnAppendEntries(request)

			select {
			case <-p.stopNotifier:
				return
			case <-p.timer.C:
				if !ok {
					p.logger.Info("timer closed")
					return
				}

				p.logState()
				p.currentState.Timeout()
			case stateName, ok := <-p.transferStateNotifier:
				if !ok {
					p.logger.Info("state transfer notifier closed")
					return
				}
				p.transferState(stateName)
			case appendEntriesResponseSender <- response:
				p.currentState.GetLogger().Debugw("send append entries response", "body", response)
			}
		case request, ok := <-voteRequestReader:
			if !ok {
				p.logger.Error(fmt.Sprintf("vote request channel of node %s closed", p.id))
				break
			}

			p.currentState.GetLogger().Debugw("receive request vote request", "term", p.currentTerm,
				"commitIndex", p.commitIndex, "lastApplied", p.lastApplied, "body", request)
			response := p.currentState.OnRequestVote(request)

			select {
			case <-p.stopNotifier:
				return
			case <-p.timer.C:
				if !ok {
					p.logger.Info("timer closed")
					return
				}

				p.logState()
				p.currentState.Timeout()
			case stateName, ok := <-p.transferStateNotifier:
				if !ok {
					p.logger.Info("state transfer notifier closed")
					return
				}
				p.transferState(stateName)
			case voteResponseSender <- response:
				p.currentState.GetLogger().Debugw("send vote response", "body", response)
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
				p.GetCurrentState().GetLogger().Errorw("commit log error", "logIndex", i, "log", p.logs[i], "err", err)
				break
			}
			p.lastApplied++
			p.GetCurrentState().GetLogger().Infow("succeed commit log", "logIndex", i, "term", p.currentTerm,
				"commitIndex", p.commitIndex, "lastApplied", p.lastApplied)
		}
	}
}

func (p *Server) transferState(state StateName) {
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
	p.currentState.GetLogger().Infow("Timeout", "term", p.currentTerm,
		"commitIndex", p.commitIndex, "lastApplied", p.lastApplied)
}

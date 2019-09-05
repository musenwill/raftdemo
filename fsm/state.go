package fsm

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type State interface {
	enterState()
	leaveState()
	onAppendEntries(param proxy.AppendEntries) proxy.Response
	onRequestVote(param proxy.RequestVote) proxy.Response
	timeout()
}

type Server struct {
	id       string
	votedFor string

	currentTerm  int64
	commitIndex  int64
	lastAplied   int64
	timer        *time.Timer
	logs         []proxy.Log
	currentState State

	commitNotifier chan bool //
	stopNotifier   chan bool // notify to stop the server

	config    *config.Config
	logger    *zap.SugaredLogger
	committer Committer
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewServer(id string, committer Committer, config *config.Config, logger *zap.SugaredLogger) *Server {
	s := &Server{
		id:             id,
		currentTerm:    0,
		commitIndex:    -1, // in raft paper, an valid index begin with 1 in no empty logs. but in program, valid index usually begin with 0
		lastAplied:     -1, // similar to commitIndex
		logs:           make([]proxy.Log, 0),
		commitNotifier: make(chan bool),
		config:         config,
		logger:         logger.With("node", id),
		committer:      committer,
	}

	s.checkConfig(config)

	// initial state is follower
	s.transferState(NewFollower(s, config))
	return s
}

func (p *Server) Run() {
	go p.commitTask()
	p.logger.Info("start up commit task")
	go p.fsmTask()
	p.logger.Info("start up fsm task")
}

func (p *Server) Stop() {
	p.transferState(NewDummyState(p, p.config))
	p.timer.Stop()

	close(p.stopNotifier)
	p.logger.Info("stop fsm task")
	close(p.commitNotifier)
	p.logger.Info("stop commit task")
}

// core code
func (p *Server) fsmTask() {
	for {
		select {
		case <-p.stopNotifier:
			return
		case <-p.timer.C:
			p.logger.Info("timeout")
			p.currentState.timeout()
		case request := <-proxy.AppendEntriesRequestReader(p.id):
			p.logger.Debugw("receive append entries request", "term", p.currentTerm, "voteFor", p.votedFor,
				"commitIndex", p.commitIndex, "lastAplied", p.lastAplied, "body", request)
			response := p.currentState.onAppendEntries(request)
			select {
			case <-p.stopNotifier:
				return
			case <-p.timer.C:
				p.logger.Info("timeout")
				p.currentState.timeout()
			case proxy.AppendEntriesResponseSender(p.id) <- response:
				p.logger.Debugw("send append entries response", "body", response)
			}
		case request := <-proxy.RequestVoteRequestReader(p.id):
			p.logger.Debugw("receive request vote request", "term", p.currentTerm, "voteFor", p.votedFor,
				"commitIndex", p.commitIndex, "lastAplied", p.lastAplied, "body", request)
			response := p.currentState.onRequestVote(request)
			select {
			case <-p.stopNotifier:
				return
			case <-p.timer.C:
				p.logger.Info("timeout")
				p.currentState.timeout()
			case proxy.RequestVoteResponseSender(p.id) <- response:
				p.logger.Debugw("send request vote response", "body", response)
			}
		}
	}
}

func (p *Server) commitTask() {
	for {
		<-p.commitNotifier
		for i := p.lastAplied + 1; i <= p.commitIndex; i++ {
			err := p.committer.Commit(p.logs[i])
			if err != nil {
				p.logger.Errorw("commit log error", "logIndex", i, "log", p.logs[i], "err", err)
				break
			}
			p.lastAplied++
			p.logger.Infow("succeed commit log", "logIndex", i)
		}
	}
}

func (p *Server) transferState(state State) {
	if p.currentState != nil {
		p.currentState.leaveState()
	}
	p.currentState = state
	p.currentState.enterState()
}

func (p *Server) checkConfig(config *config.Config) {
	if config.Timeout <= 0 {
		panic(fmt.Sprintf("invalid timeout %v, expected greater than 0", config.Timeout))
	}
	if config.Timeout < 20 {
		fmt.Printf("timeout too small %v, may cause a lot of election, expected greater than 20ms", config.Timeout)
	}
	if config.Timeout > 1000 {
		fmt.Printf("timeout is %v, may cause performance quite slow, are you sure", config.Timeout)
	}
	if len(config.Nodes) <= 0 {
		panic("empty nodes")
	}
	set := make(map[string]bool)
	for _, id := range config.Nodes {
		if _, ok := set[id.ID]; ok {
			panic("duplicate nodes")
		} else {
			set[id.ID] = true
		}
	}
}

func (p *Server) getCommitIndex() int64 {
	return atomic.LoadInt64(&p.commitIndex)
}

func (p *Server) setCommitIndex(i int64) {
	atomic.StoreInt64(&p.commitIndex, i)
	p.commitNotifier <- true
}

func (p *Server) lastLogIndex() int64 {
	return int64(len(p.logs) - 1)
}

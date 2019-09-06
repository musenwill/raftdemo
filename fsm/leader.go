package fsm

import (
	"context"
	"sync"
	"time"

	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
)

type nodeIndex struct {
	nextIndex, matchIndex int64
}

type Leader struct {
	*Server       // embed server
	nIndex        map[string]*nodeIndex
	stopReplicate chan bool
	stateLogger   *zap.SugaredLogger

	stopReplicateLock *sync.Mutex
}

func NewLeader(s *Server, conf *config.Config) *Leader {
	lastLogIndex := s.lastLogIndex()
	nIndex := make(map[string]*nodeIndex)
	for _, n := range conf.Nodes {
		if n.ID != s.id {
			nIndex[n.ID] = &nodeIndex{nextIndex: lastLogIndex + 1, matchIndex: 0}
		}
	}
	return &Leader{
		Server:            s,
		nIndex:            nIndex,
		stopReplicate:     nil,
		stateLogger:       s.logger.With("state", "leader"),
		stopReplicateLock: &sync.Mutex{},
	}
}

func (p *Leader) implStateInterface() {
	var _ State = &Leader{}
}

func (p *Leader) getLogger() *zap.SugaredLogger {
	return p.stateLogger
}

func (p *Leader) initState() {
	p.rapidResetTimer()
	p.resetStopReplicate()
}

func (p *Leader) enterState() {
	p.stateLogger.Infow("enter state", "term", p.currentTerm, "voteFor", p.votedFor,
		"commitIndex", p.commitIndex, "lastAplied", p.lastAplied)
	p.initState()
	p.heartbeatJob()
}

func (p *Leader) leaveState() {
	p.resetStopReplicate()
	p.stateLogger.Infow("leave state", "term", p.currentTerm, "voteFor", p.votedFor,
		"commitIndex", p.commitIndex, "lastAplied", p.lastAplied)
}

func (p *Leader) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	if param.Term <= p.getCurrentTerm() {
		return proxy.Response{Term: p.getCurrentTerm(), Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.onAppendEntries(param)
	}
}

func (p *Leader) onRequestVote(param proxy.RequestVote) proxy.Response {
	if param.Term <= p.getCurrentTerm() {
		return proxy.Response{Term: p.getCurrentTerm(), Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.onRequestVote(param)
	}
}

func (p *Leader) timeout() {
	p.initState()
	p.replicationJob()
}

func (p *Leader) heartbeatJob() {
	p.job(p.heartbeat)
}

func (p *Leader) replicationJob() {
	p.job(p.replicate)
}

type jobHandler func(ctx context.Context, nodeID string)

func (p *Leader) job(handler jobHandler) {
	ctx, cancel := context.WithCancel(context.Background())
	// make sure jobs are canceled when timeout
	go func() {
		<-p.stopReplicate
		cancel()
	}()

	for _, node := range p.config.Nodes {
		if node.ID == p.id {
			continue
		}
		select {
		case <-p.stopReplicate:
			return
		default:
			go handler(ctx, node.ID)
		}
	}
}

func (p *Leader) heartbeat(ctx context.Context, nodeID string) {
	response, err := proxy.SendAppendEntries(ctx, nodeID,
		proxy.AppendEntries{
			Term:         p.getCurrentTerm(),
			LeaderID:     p.id,
			LeaderCommit: p.getCommitIndex()})
	if err != nil {
		// log it
		return
	}
	if response.Term > p.getCurrentTerm() {
		p.transferState(NewFollower(p.Server, p.config))
		return
	}
}

// @TODO: to be optimized
func (p *Leader) replicate(ctx context.Context, nodeID string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			{
				nextLogIndex := p.nIndex[nodeID].nextIndex
				preLogIndex := nextLogIndex - 1
				replicationBound := nextLogIndex + int64(p.config.MaxReplicate)
				if replicationBound > p.lastLogIndex()+1 {
					replicationBound = p.lastLogIndex() + 1
				}
				var preLogTerm int64 = 0
				if preLogIndex >= 0 {
					preLogTerm = p.logs[preLogIndex].Term
				}

				response, err := proxy.SendAppendEntries(ctx, nodeID,
					proxy.AppendEntries{
						Term:         p.getCurrentTerm(),
						LeaderID:     p.id,
						PrevLogIndex: preLogIndex,
						PrevLogTerm:  preLogTerm,
						LeaderCommit: p.getCommitIndex(),
						Entries:      p.logs[nextLogIndex:replicationBound],
					})
				if err != nil {
					// log it
					return
				}
				if response.Term > p.getCurrentTerm() {
					p.transferState(NewFollower(p.Server, p.config))
					return
				}
				// update commit index
				if response.Success {
					p.nIndex[nodeID].nextIndex = replicationBound
					p.nIndex[nodeID].matchIndex = replicationBound - 1

					for N := p.lastLogIndex(); N > p.getCommitIndex(); N-- {
						if p.logs[N].Term < p.getCurrentTerm() {
							break
						}
						count := 1 // add self in firstly
						for _, v := range p.nIndex {
							if v.matchIndex >= N {
								count++
							}
						}
						if count > p.config.Len()/2 {
							p.setCommitIndex(N)
						}
					}
					return
				}

				// if failed, just retry util reach bottom
				p.nIndex[nodeID].nextIndex--
				if p.nIndex[nodeID].nextIndex < 0 {
					p.nIndex[nodeID].nextIndex = 0
					return
				}
			}
		}
	}
}

// reset timer for leader, which should be a bit faster than the timer of follower
// to avoid follower timeout
func (p *Server) rapidResetTimer() {
	rapidTimer := int(float64(p.config.Timeout) * 0.8)

	p.timerLock.Lock()
	defer p.timerLock.Unlock()

	if p.timer == nil {
		p.timer = time.NewTimer(time.Duration(rapidTimer) * time.Millisecond)
	} else {
		p.timer.Reset(time.Duration(rapidTimer) * time.Millisecond)
	}
}

func (p *Leader) resetStopReplicate() {
	p.stopReplicateLock.Lock()
	defer p.stopReplicateLock.Unlock()

	if p.stopReplicate != nil {
		close(p.stopReplicate)
	}
	p.stopReplicate = make(chan bool)
}

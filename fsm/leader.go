package fsm

import (
	"fmt"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"go.uber.org/zap"
	"sync"
)

type nodeIndex struct {
	nextIndex, matchIndex int64
}

type Leader struct {
	Prober
	nIndex        map[string]*nodeIndex
	stopReplicate chan bool
	stateLogger   *zap.SugaredLogger
}

func NewLeader(s Prober, logger *zap.SugaredLogger, nodes []model.Node) *Leader {
	lastLogIndex := s.GetLastLogIndex()
	nIndex := make(map[string]*nodeIndex)
	hostID := s.GetHost()

	for _, n := range nodes {
		if n.ID != hostID {
			nIndex[n.ID] = &nodeIndex{nextIndex: lastLogIndex + 1, matchIndex: 0}
		}
	}
	return &Leader{
		s, nIndex, nil, logger.With("state", StateEnum.Leader),
	}
}

func (p *Leader) implStateInterface() {
	var _ State = &Leader{}
}

func (p *Leader) GetLogger() *zap.SugaredLogger {
	return p.stateLogger
}

func (p *Leader) EnterState() {
	p.stateLogger.Infow("enter state", "term", p.GetTerm(),
		"commitIndex", p.GetCommitIndex(), "lastApplied", p.GetLastAppliedIndex())
	p.rapidResetTimer()
}

func (p *Leader) LeaveState() {
	p.resetStopReplicate()
	p.stateLogger.Infow("leave state", "term", p.GetTerm(),
		"commitIndex", p.GetCommitIndex(), "lastApplied", p.GetLastAppliedIndex())
}

func (p *Leader) OnAppendEntries(param proxy.AppendEntries) proxy.Response {
	term := p.GetTerm()
	if param.Term <= term {
		return proxy.Response{Term: term, Success: false}
	} else {
		p.NotifyTransferState(StateEnum.Follower)
		return p.GetCurrentState().OnAppendEntries(param)
	}
}

func (p *Leader) OnRequestVote(param proxy.RequestVote) proxy.Response {
	term := p.GetTerm()
	if param.Term <= term {
		return proxy.Response{Term: term, Success: false}
	} else {
		p.NotifyTransferState(StateEnum.Follower)
		return p.GetCurrentState().OnRequestVote(param)
	}
}

func (p *Leader) Timeout() {
	p.EnterState()
	p.replicateJob()
}

func (p *Leader) replicateJob() {
	wg := &sync.WaitGroup{}
	wg.Add(p.GetConfig().GetNodeCount() - 1)
	hostID := p.GetHost()

	for _, n := range p.GetConfig().GetNodes() {
		node := n
		if node.ID == hostID {
			continue
		}
		go func() {
			defer wg.Done()

			select {
			case <-p.stopReplicate:
				return
			default:
				p.replicate(node.ID)
			}
		}()
	}

	wg.Wait()
}

func (p *Leader) replicate(nodeID string) {
	for {
		lastLogIndex := p.GetLastLogIndex()
		nextLogIndex := p.nIndex[nodeID].nextIndex
		preLogIndex := nextLogIndex - 1
		replicationBound := nextLogIndex + int64(p.GetConfig().GetReplicateUnitSize())
		if replicationBound > lastLogIndex+1 {
			replicationBound = lastLogIndex + 1
		}
		var preLogTerm int64 = 0
		if preLogIndex >= 0 {
			preLogTerm = p.GetLog(preLogIndex).Term
		}

		request := proxy.AppendEntries{
			Term:         p.GetTerm(),
			LeaderID:     p.GetHost(),
			PrevLogIndex: preLogIndex,
			PrevLogTerm:  preLogTerm,
			LeaderCommit: p.GetCommitIndex(),
			Entries:      p.GetLogs()[nextLogIndex:replicationBound],
		}

		appendIndexRequestSender, err := p.GetProxy().AppendEntriesRequestSender(nodeID)
		if err != nil {
			p.stateLogger.Error(err)
			return
		}
		appendIndexResponseReader, err := p.GetProxy().AppendEntriesResponseReader(nodeID)
		if err != nil {
			p.stateLogger.Error(err)
			return
		}

		select {
		case <-p.stopReplicate:
			return
		case appendIndexRequestSender <- request:
			select {
			case <-p.stopReplicate:
				return
			case response, ok := <-appendIndexResponseReader:
				if !ok {
					p.stateLogger.Error(fmt.Sprintf("append index response channel of node %s closed", nodeID))
					return
				}

				if response.Term > p.GetTerm() {
					p.NotifyTransferState(StateEnum.Follower)
					return
				}
				// update commit index
				if response.Success {
					p.updateCommitIndex(nodeID, replicationBound)
					return
				}
				// if failed, just retry util reach bottom
				p.nIndex[nodeID].nextIndex--
				if p.nIndex[nodeID].nextIndex < 0 {
					p.nIndex[nodeID].nextIndex = 0
				}
				return
			}
		}
	}
}

// reset timer for leader, which should be a bit faster than the timer of follower
// to avoid follower Timeout
func (p *Leader) rapidResetTimer() {
	rapidTimer := int64(float64(p.GetConfig().GetReplicateTimeout()) * 0.8)
	p.SetTimer(rapidTimer)
	p.resetStopReplicate()
}

func (p *Leader) resetStopReplicate() {
	if p.stopReplicate != nil {
		close(p.stopReplicate)
	}
	p.stopReplicate = make(chan bool)
}

func (p *Leader) updateCommitIndex(nodeID string, replicationBound int64) {
	p.nIndex[nodeID].nextIndex = replicationBound
	p.nIndex[nodeID].matchIndex = replicationBound - 1

	term := p.GetTerm()
	commitIndex := p.GetCommitIndex()
	for N := p.GetLastLogIndex(); N > commitIndex; N-- {
		if p.GetLog(N).Term < term {
			break
		}
		count := 1 // add self in firstly
		for _, v := range p.nIndex {
			if v.matchIndex >= N {
				count++
			}
		}
		if count > p.GetConfig().GetNodeCount()/2 {
			p.SetCommitIndex(N)
		}
	}
}

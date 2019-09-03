package fsm

import (
	"context"
	"sync"

	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
)

type nodeIndex struct {
	nextIndex, matchIndex int64
}

type Leader struct {
	*Server       // embed server
	nIndex        map[string]*nodeIndex
	stopReplicate chan bool
}

func NewLeader(s *Server, conf *config.Config) *Leader {
	lastLogIndex := s.lastLogIndex()
	nIndex := make(map[string]*nodeIndex)
	for _, n := range conf.Nodes {
		if n.ID != s.id {
			nIndex[n.ID] = &nodeIndex{nextIndex: lastLogIndex + 1, matchIndex: 0}
		}
	}
	return &Leader{s, nIndex, nil}
}

func (p *Leader) implStateInterface() {
	var _ State = &Leader{}
}

func (p *Leader) initState() {
	p.resetTimer()
	if p.stopReplicate != nil {
		close(p.stopReplicate)
	}
	p.stopReplicate = make(chan bool)
}

func (p *Leader) enterState() {
	p.initState()
	p.heartbeatJob()
}

func (p *Leader) leaveState() {
	if p.stopReplicate != nil {
		close(p.stopReplicate)
	}
	p.stopReplicate = nil
}

func (p *Leader) onAppendEntries(param proxy.AppendEntries) proxy.Response {
	if param.Term <= p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.onAppendEntries(param)
	}
}

func (p *Leader) onRequestVote(param proxy.RequestVote) proxy.Response {
	if param.Term <= p.currentTerm {
		return proxy.Response{Term: p.currentTerm, Success: false}
	} else {
		p.transferState(NewFollower(p.Server, p.config))
		return p.onRequestVote(param)
	}
}

func (p *Leader) timeout() {
	p.resetTimer()
}

func (p *Leader) heartbeatJob() {
	p.job(p.heartbeat)
}

type jobHandler func(ctx context.Context, wg *sync.WaitGroup, nodeID string)

func (p *Leader) job(handler jobHandler) {
	ctx, cancel := context.WithCancel(context.Background())
	// make sure jobs are canceled when timeout
	go func() {
		<-p.stopReplicate
		cancel()
	}()

	wg := &sync.WaitGroup{}
	wg.Add(len(p.config.Nodes))

	for _, node := range p.config.Nodes {
		select {
		case <-p.stopReplicate:
			return
		default:
			go handler(ctx, wg, node.ID)
		}
	}

	wg.Wait()
}

func (p *Leader) heartbeat(ctx context.Context, wg *sync.WaitGroup, nodeID string) {
	defer wg.Done()

	response, err := proxy.SendAppendEntries(ctx, nodeID,
		proxy.AppendEntries{
			Term:         p.currentTerm,
			LeaderID:     p.id,
			LeaderCommit: p.getCommitIndex()})
	if err != nil {
		// log it
		return
	}
	if response.Term > p.currentTerm {
		p.transferState(NewFollower(p.Server, p.config))
		return
	}
}

func (p *Leader) replicate(ctx context.Context, wg *sync.WaitGroup, nodeID string) {
	defer wg.Done()

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

				response, err := proxy.SendAppendEntries(ctx, nodeID,
					proxy.AppendEntries{
						Term:         p.currentTerm,
						LeaderID:     p.id,
						PrevLogIndex: preLogIndex,
						PrevLogTerm:  p.logs[preLogIndex].Term,
						LeaderCommit: p.getCommitIndex(),
						Entries:      p.logs[nextLogIndex:replicationBound],
					})
				if err != nil {
					// log it
					return
				}
				if response.Term > p.currentTerm {
					p.transferState(NewFollower(p.Server, p.config))
					return
				}
				// update commit index
				if response.Success {
					p.nIndex[nodeID].nextIndex = replicationBound
					p.nIndex[nodeID].matchIndex = replicationBound - 1

					for N := p.lastLogIndex(); N > p.getCommitIndex(); N-- {
						if p.logs[N].Term < p.currentTerm {
							break
						}
						count := 1 // add self in firstly
						for _, v := range p.nIndex {
							if v.matchIndex >= N {
								count++
							}
						}
						if count > len(p.config.Nodes)/2 {
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

package proxy

import (
	"fmt"
	"github.com/musenwill/raftdemo/model"
)

type ChanProxy struct {
	router map[string]endpoint
}

func NewChanProxy() *ChanProxy {
	return &ChanProxy{}
}

func (p *ChanProxy) implProxy() {
	var _ Proxy = &ChanProxy{}
}

func (p *ChanProxy) AppendEntriesRequestSender(nodeID string) (chan<- AppendEntries, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.appendEntries.request, nil
}

func (p *ChanProxy) AppendEntriesRequestReader(nodeID string) (<-chan AppendEntries, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.appendEntries.request, nil
}

func (p *ChanProxy) AppendEntriesResponseSender(nodeID string) (chan<- Response, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.appendEntries.response, nil
}

func (p *ChanProxy) AppendEntriesResponseReader(nodeID string) (<-chan Response, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.appendEntries.response, nil
}

func (p *ChanProxy) VoteRequestSender(nodeID string) (chan<- RequestVote, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.requestVote.request, nil
}

func (p *ChanProxy) VoteRequestReader(nodeID string) (<-chan RequestVote, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.requestVote.request, nil
}

func (p *ChanProxy) VoteResponseSender(nodeID string) (chan<- Response, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.requestVote.response, nil
}

func (p *ChanProxy) VoteResponseReader(nodeID string) (<-chan Response, error) {
	e, ok := p.router[nodeID]
	if !ok {
		return nil, fmt.Errorf("node with id %v not exist", nodeID)
	}
	return e.requestVote.response, nil
}

/* register nodes to proxy, add new nodes and remove those not already exists */
func (p *ChanProxy) Register(nodes []model.Node) {
	if p.router == nil {
		p.router = make(map[string]endpoint)
	}

	nodeSet := make(map[string]bool)
	for _, node := range nodes {
		nodeSet[node.ID] = true
	}

	// clear all none exists nodes
	for k, v := range p.router {
		if !nodeSet[k] {
			close(v.appendEntries.request)
			close(v.appendEntries.response)
			close(v.requestVote.request)
			close(v.requestVote.response)
			delete(p.router, k)
		}
	}

	// add all new joined nodes
	for k := range nodeSet {
		if _, ok := p.router[k]; !ok {
			e := endpoint{
				appendEntries: appendEntriesEndpoint{
					request:  make(chan AppendEntries),
					response: make(chan Response),
				},
				requestVote: requestVoteEndpoint{
					request:  make(chan RequestVote),
					response: make(chan Response),
				},
			}
			p.router[k] = e
		}
	}
}

type appendEntriesEndpoint struct {
	request  chan AppendEntries
	response chan Response
}

type requestVoteEndpoint struct {
	request  chan RequestVote
	response chan Response
}

type endpoint struct {
	appendEntries appendEntriesEndpoint
	requestVote   requestVoteEndpoint
}

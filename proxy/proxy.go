package proxy

import (
	"fmt"
	"time"

	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
)

type ErrProxyAbort struct {
	err error
}

func (e *ErrProxyAbort) Error() string {
	return e.err.Error()
}

type ChanProxy struct {
	router  map[string]endpoint
	timeout time.Duration
}

func NewChanProxy(nodes []raft.Node, timeout time.Duration) *ChanProxy {
	proxy := &ChanProxy{
		router:  make(map[string]endpoint),
		timeout: timeout,
	}

	for _, node := range nodes {
		if _, ok := proxy.router[node.ID]; !ok {
			e := endpoint{
				appendEntries: appendEntriesEndpoint{
					request:  make(chan model.AppendEntries, 1),
					response: make(chan model.Response, 1),
				},
				requestVote: requestVoteEndpoint{
					request:  make(chan model.RequestVote, 1),
					response: make(chan model.Response, 1),
				},
			}
			proxy.router[node.ID] = e
		}
	}

	return proxy
}

func (c *ChanProxy) Send(nodeID string, request interface{}, abort chan bool) (model.Response, error) {
	e, ok := c.router[nodeID]
	if !ok {
		return model.Response{}, fmt.Errorf("node with id %v not exist", nodeID)
	}
	timeout := time.NewTimer(c.timeout)

	switch t := request.(type) {
	case model.AppendEntries:
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("send append entries to %s aborted", nodeID)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("send append entries to %s timeout", nodeID)
		case e.appendEntries.request <- request.(model.AppendEntries):
		}
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("receive append entries response from %s aborted", nodeID)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("receive append entries response from %s timeout", nodeID)
		case response := <-e.appendEntries.response:
			return response, nil
		}
	case model.RequestVote:
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("send request vote to %s aborted", nodeID)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("send request vote to %s timeout", nodeID)
		case e.requestVote.request <- request.(model.RequestVote):
		}
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("receive request vote response from %s aborted", nodeID)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("recevie request vote response from %s timeout", nodeID)
		case response := <-e.requestVote.response:
			return response, nil
		}
	default:
		return model.Response{}, fmt.Errorf("unknown request type to %s: %v", nodeID, t)
	}
}

func (c *ChanProxy) Receive(nodeID string, f func(request interface{}) model.Response, abort chan bool) error {
	e, ok := c.router[nodeID]
	if !ok {
		return fmt.Errorf("node with id %v not exist", nodeID)
	}

	select {
	case <-abort:
		return &ErrProxyAbort{fmt.Errorf("aborted")}
	case request := <-e.appendEntries.request:
		select {
		case <-abort:
			return &ErrProxyAbort{fmt.Errorf("aborted")}
		case e.appendEntries.response <- f(request):
		}
	case request := <-e.requestVote.request:
		select {
		case <-abort:
			return &ErrProxyAbort{fmt.Errorf("aborted")}
		case e.requestVote.response <- f(request):
		}
	}

	return nil
}

type appendEntriesEndpoint struct {
	request  chan model.AppendEntries
	response chan model.Response
}

type requestVoteEndpoint struct {
	request  chan model.RequestVote
	response chan model.Response
}

type endpoint struct {
	appendEntries appendEntriesEndpoint
	requestVote   requestVoteEndpoint
}

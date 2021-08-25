package proxy

import (
	"fmt"
	"time"

	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
)

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
					request:  make(chan model.AppendEntries),
					response: make(chan model.Response),
				},
				requestVote: requestVoteEndpoint{
					request:  make(chan model.RequestVote),
					response: make(chan model.Response),
				},
			}
			proxy.router[node.ID] = e
		}
	}

	return proxy
}

func (p *ChanProxy) ensureImplProxy() {
	var _ Proxy = &ChanProxy{}
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
			return model.Response{}, fmt.Errorf("aborted")
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("timeout")
		case e.appendEntries.request <- request.(model.AppendEntries):
		}
		select {
		case <-abort:
			return model.Response{}, fmt.Errorf("aborted")
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("timeout")
		case response := <-e.appendEntries.response:
			return response, nil
		}
	case model.RequestVote:
		select {
		case <-abort:
			return model.Response{}, fmt.Errorf("aborted")
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("timeout")
		case e.requestVote.request <- request.(model.RequestVote):
		}
		select {
		case <-abort:
			return model.Response{}, fmt.Errorf("aborted")
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("timeout")
		case response := <-e.requestVote.response:
			return response, nil
		}
	default:
		return model.Response{}, fmt.Errorf("unknown request type: %v", t)
	}
}

func (c *ChanProxy) Receive(nodeID string, f func(request interface{}) model.Response, abort chan bool) error {
	e, ok := c.router[nodeID]
	if !ok {
		return fmt.Errorf("node with id %v not exist", nodeID)
	}

	select {
	case <-abort:
		return fmt.Errorf("aborted")
	case request := <-e.appendEntries.request:
		e.appendEntries.response <- f(request)
	case request := <-e.requestVote.request:
		e.requestVote.response <- f(request)
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

package proxy

import (
	"fmt"
	"sync"
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

type ErrPipeBroken struct {
	err error
}

func (e *ErrPipeBroken) Error() string {
	return e.err.Error()
}

type ChanProxy struct {
	router      map[string]endpoint
	timeout     time.Duration
	brokenPipes map[string]map[string]bool
	sync.RWMutex
}

func NewChanProxy(nodes []raft.Node, timeout time.Duration) *ChanProxy {
	proxy := &ChanProxy{
		router:      make(map[string]endpoint),
		timeout:     timeout,
		brokenPipes: make(map[string]map[string]bool),
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

func (c *ChanProxy) BrokenPipes() []model.Pipe {
	pipes := make([]model.Pipe, 0)

	c.RLock()
	defer c.RUnlock()

	for from, v := range c.brokenPipes {
		for to := range v {
			pipes = append(pipes, model.Pipe{From: from, To: to, State: model.StatePipe_broken})
		}
	}
	return pipes
}

func (c *ChanProxy) SetPipeStates(pipes []model.Pipe) error {
	for _, p := range pipes {
		_, okFrom := c.router[p.From]
		_, okTo := c.router[p.To]
		if !okFrom {
			return fmt.Errorf("node %s not exists", p.From)
		}
		if !okTo {
			return fmt.Errorf("node %s not exists", p.To)
		}
	}

	c.Lock()
	defer c.Unlock()

	for _, p := range pipes {
		if p.State == model.StatePipe_broken {
			v := c.brokenPipes[p.From]
			if v == nil {
				v = make(map[string]bool)
				c.brokenPipes[p.From] = v
			}
			v[p.To] = true
		} else {
			v := c.brokenPipes[p.From]
			if v != nil && v[p.To] {
				delete(v, p.To)
			}
		}
	}

	return nil
}

func (c *ChanProxy) CheckPipeBroken(from, to string) error {
	c.RLock()
	defer c.RUnlock()

	v := c.brokenPipes[from]
	if v == nil {
		return nil
	}

	if v[to] {
		return &ErrPipeBroken{fmt.Errorf("pipe from %s to %s is broken", from, to)}
	}

	return nil
}

func (c *ChanProxy) Send(from, to string, request interface{}, abort chan bool) (model.Response, error) {
	if err := c.CheckPipeBroken(from, to); err != nil {
		return model.Response{}, err
	}

	e, ok := c.router[to]
	if !ok {
		return model.Response{}, fmt.Errorf("node with id %v not exist", to)
	}
	timeout := time.NewTimer(c.timeout)

	switch t := request.(type) {
	case model.AppendEntries:
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("send append entries to %s aborted", to)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("send append entries to %s timeout", to)
		case e.appendEntries.request <- request.(model.AppendEntries):
		}
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("receive append entries response from %s aborted", to)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("receive append entries response from %s timeout", to)
		case response := <-e.appendEntries.response:
			return response, nil
		}
	case model.RequestVote:
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("send request vote to %s aborted", to)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("send request vote to %s timeout", to)
		case e.requestVote.request <- request.(model.RequestVote):
		}
		select {
		case <-abort:
			return model.Response{}, &ErrProxyAbort{fmt.Errorf("receive request vote response from %s aborted", to)}
		case <-timeout.C:
			return model.Response{}, fmt.Errorf("recevie request vote response from %s timeout", to)
		case response := <-e.requestVote.response:
			return response, nil
		}
	default:
		return model.Response{}, fmt.Errorf("unknown request type to %s: %v", to, t)
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

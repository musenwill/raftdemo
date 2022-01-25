package proxy

import (
	"github.com/musenwill/raftdemo/model"
)

type Proxy interface {
	Send(from, to string, request interface{}, abort chan bool) (model.Response, error)
	Receive(nodeID string, f func(request interface{}) model.Response, abort chan bool) error
	BrokenPipes() []model.Pipe
	SetPipeStates(pipes []model.Pipe) error
}

package proxy

import (
	"github.com/musenwill/raftdemo/model"
)

type Proxy interface {
	Send(nodeID string, request interface{}, abort chan bool) (model.Response, error)
	Receive(nodeID string, f func(request interface{}) model.Response, abort chan bool) error
}

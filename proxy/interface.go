package proxy

import (
	"github.com/musenwill/raftdemo/model"
)

type Proxy interface {
	Send(nodeID string, request interface{}) (model.Response, error)
	Receive(nodeID string, f func(request interface{}) error) error
}

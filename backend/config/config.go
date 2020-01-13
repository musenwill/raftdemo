package config

import (
	"errors"
	"fmt"
	"github.com/musenwill/raftdemo/model"
	"sync"
)

func NewDefaultConfig(nodes []model.Node, replicateUnitSize int64, replicateTimeout int64, maxLogSize int64) *DefaultConfig {
	conf := &DefaultConfig{
		NodesLock:             &sync.RWMutex{},
		ReplicateTimeoutLock:  &sync.RWMutex{},
		ReplicateUnitSizeLock: &sync.RWMutex{},
		MaxLogSizeLock:        &sync.RWMutex{},
	}
	conf.SetNodes(nodes)
	conf.SetReplicateTimeout(replicateTimeout)
	conf.SetReplicateUnitSize(replicateUnitSize)
	conf.SetMaxLogSize(maxLogSize)

	return conf
}

type DefaultConfig struct {
	Nodes     []model.Node
	NodesLock *sync.RWMutex

	ReplicateTimeout     int64
	ReplicateTimeoutLock *sync.RWMutex

	ReplicateUnitSize     int64
	ReplicateUnitSizeLock *sync.RWMutex

	MaxLogSize     int64
	MaxLogSizeLock *sync.RWMutex
}

func (p *DefaultConfig) ensureImplConfig() {
	var _ Config = &DefaultConfig{}
}

func (p *DefaultConfig) GetNodes() []model.Node {
	p.NodesLock.RLock()
	defer p.NodesLock.RUnlock()

	cpy := make([]model.Node, len(p.Nodes))
	copy(cpy, p.Nodes)

	return cpy
}

func (p *DefaultConfig) SetNodes(nodes []model.Node) {
	p.NodesLock.Lock()
	defer p.NodesLock.Unlock()

	cpy := make([]model.Node, len(nodes))
	copy(cpy, nodes)

	p.Nodes = cpy
}

func (p *DefaultConfig) GetNodeCount() int {
	p.NodesLock.RLock()
	defer p.NodesLock.RUnlock()

	return len(p.Nodes)
}

func (p *DefaultConfig) GetReplicateUnitSize() int64 {
	p.ReplicateUnitSizeLock.RLock()
	defer p.ReplicateUnitSizeLock.RUnlock()

	return p.ReplicateUnitSize
}

func (p *DefaultConfig) SetReplicateUnitSize(unitSize int64) {
	p.ReplicateUnitSizeLock.Lock()
	defer p.ReplicateUnitSizeLock.Unlock()

	p.ReplicateUnitSize = unitSize
}

func (p *DefaultConfig) GetReplicateTimeout() int64 {
	p.ReplicateTimeoutLock.RLock()
	defer p.ReplicateTimeoutLock.RUnlock()

	return p.ReplicateTimeout
}

func (p *DefaultConfig) SetReplicateTimeout(timeout int64) {
	p.ReplicateTimeoutLock.Lock()
	defer p.ReplicateTimeoutLock.Unlock()

	p.ReplicateTimeout = timeout
}

func (p *DefaultConfig) GetMaxLogSize() int64 {
	p.MaxLogSizeLock.RLock()
	defer p.MaxLogSizeLock.RUnlock()

	return p.MaxLogSize
}

func (p *DefaultConfig) SetMaxLogSize(maxLogSize int64) {
	p.MaxLogSizeLock.Lock()
	defer p.MaxLogSizeLock.Unlock()

	p.MaxLogSize = maxLogSize
}

func (p *DefaultConfig) Check() error {
	if p.GetReplicateTimeout() <= 0 {
		return errors.New(fmt.Sprintf("invalid Timeout %v, expected greater than 0", p.GetReplicateTimeout()))
	}
	if p.GetReplicateTimeout() < 20 {
		fmt.Printf("Timeout too small %v, may cause a lot of election, expected greater than 20ms", p.GetReplicateTimeout())
	}
	if p.GetReplicateTimeout() > 1000 {
		fmt.Printf("Timeout is %v, may cause performance quite slow", p.GetReplicateTimeout())
	}
	if p.GetNodeCount() <= 0 {
		return errors.New("empty nodes")
	}
	set := make(map[string]bool)
	for _, id := range p.GetNodes() {
		if set[id.ID] {
			return errors.New("duplicate nodes")
		} else {
			set[id.ID] = true
		}
	}
	return nil
}

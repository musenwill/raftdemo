package config

import (
	"fmt"
	"github.com/musenwill/raftdemo/model"
	"sync"
)

func NewDefaultConfig(nodes []model.Node, replicateUnitSize int, replicateTimeout int64) *DefaultConfig {
	return &DefaultConfig{
		Nodes:                 nodes,
		NodesLock:             &sync.RWMutex{},
		ReplicateTimeout:      replicateTimeout,
		ReplicateTimeoutLock:  &sync.RWMutex{},
		ReplicateUnitSize:     replicateUnitSize,
		ReplicateUnitSizeLock: &sync.RWMutex{},
	}
}

type DefaultConfig struct {
	Nodes     []model.Node
	NodesLock *sync.RWMutex

	ReplicateTimeout     int64 // milliseconds
	ReplicateTimeoutLock *sync.RWMutex

	ReplicateUnitSize     int
	ReplicateUnitSizeLock *sync.RWMutex
}

func (p *DefaultConfig) ensureImplConfig() {
	var _ Config = &DefaultConfig{}
}

func (p *DefaultConfig) GetNodes() []model.Node {
	p.NodesLock.RLock()
	defer p.NodesLock.RUnlock()

	return p.Nodes[:]
}

func (p *DefaultConfig) SetNodes(nodes []model.Node) {
	p.NodesLock.Lock()
	defer p.NodesLock.Unlock()

	p.Nodes = nodes[:]
}

func (p *DefaultConfig) GetNodeCount() int {
	p.NodesLock.RLock()
	defer p.NodesLock.RUnlock()

	return len(p.Nodes)
}

func (p *DefaultConfig) GetReplicateUnitSize() int {
	p.ReplicateUnitSizeLock.RLock()
	defer p.ReplicateUnitSizeLock.RUnlock()

	return p.ReplicateUnitSize
}

func (p *DefaultConfig) SetReplicateUnitSize(unitSize int) {
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

func (p *DefaultConfig) Check() {
	if p.GetReplicateTimeout() <= 0 {
		panic(fmt.Sprintf("invalid Timeout %v, expected greater than 0", p.GetReplicateTimeout()))
	}
	if p.GetReplicateTimeout() < 20 {
		fmt.Printf("Timeout too small %v, may cause a lot of election, expected greater than 20ms", p.GetReplicateTimeout())
	}
	if p.GetReplicateTimeout() > 1000 {
		fmt.Printf("Timeout is %v, may cause performance quite slow, are you sure", p.GetReplicateTimeout())
	}
	if p.GetNodeCount() <= 0 {
		panic("empty nodes")
	}
	set := make(map[string]bool)
	for _, id := range p.GetNodes() {
		if set[id.ID] {
			panic("duplicate nodes")
		} else {
			set[id.ID] = true
		}
	}
}

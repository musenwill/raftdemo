package config

import "github.com/musenwill/raftdemo/model"

type Config interface {
	GetNodes() []model.Node

	SetNodes(nodes []model.Node)

	GetNodeCount() int

	GetReplicateUnitSize() int

	SetReplicateUnitSize(int)

	GetReplicateTimeout() int64

	SetReplicateTimeout(timeout int64)
}

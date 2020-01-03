package config

import (
	"github.com/musenwill/raftdemo/model"
)

type Config interface {
	GetNodes() []model.Node

	SetNodes(nodes []model.Node)

	GetNodeCount() int

	GetMaxLogSize() int64 // in bytes

	SetMaxLogSize(int64)

	GetReplicateUnitSize() int64

	SetReplicateUnitSize(int64)

	GetReplicateTimeout() int64 // in milliseconds

	SetReplicateTimeout(timeout int64)
}

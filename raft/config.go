package raft

import (
	"time"

	"github.com/musenwill/raftdemo/log"
	"go.uber.org/zap"
)

type Config struct {
	ElectionTimeout  time.Duration `json:"election_timeout"`
	ReplicateTimeout time.Duration `json:"replication_timeout"`
	CampaignTimeout  time.Duration `json:"compaign_timeout"`
	ElectionRandom   time.Duration `json:"election_random"`
	RequestTimeout   time.Duration `json:"request_timeout"`

	Nodes             uint `json:"nodes"`
	ReplicateMaxBatch uint `json:"replicate_max_batch"` // 0 for no limit
	MaxDataSize       uint `json:"max_data_size"`       // max data size can receive, 0 for no limit

	LogFile  string       `json:"log_file"`
	LogLevel log.LogLevel `json:"log_level"`
	Logger   *zap.Logger  `json:""`
}

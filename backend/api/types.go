package api

import (
	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/fsm"
	"github.com/musenwill/raftdemo/model"
	proxy2 "github.com/musenwill/raftdemo/proxy"
	"strings"
)

type Node struct {
	Host          string `json:"host"`
	Term          int64  `json:"term"`
	State         string `json:"state"`
	CommitIndex   int64  `json:"commit_index"`
	LastAppliedID int64  `json:"last_applied_id"`

	Leader  string `json:"leader"`
	VoteFor string `json:"vote_for"`

	Logs []model.Log `json:"logs"`
}

type ListResponse struct {
	Total   int           `json:"total"`
	Entries []interface{} `json:"entries"`
}

type Context struct {
	NodeMap map[string]fsm.Prober
	Proxy   *proxy2.ChanProxy
	Conf    config.Config
	Logger  *common.Logger
}

type ConfigInfo struct {
	LogLevel          string   `json:"log_level"`
	Nodes             []string `json:"nodes"`
	ReplicateTimeout  int64    `json:"replicate_timeout"`
	ReplicateUnitSize int64    `json:"replicate_unit_size"`
	MaxLogSize        int64    `json:"max_log_size"`
}

type NodeList []Node

func (p NodeList) Len() int {
	return len(p)
}

func (p NodeList) Less(i, j int) bool {
	return strings.Compare(p[i].Host, p[j].Host) < 0
}

func (p NodeList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

package config

import (
	"github.com/musenwill/raftdemo/log"

	"github.com/musenwill/raftdemo/http"
	"github.com/musenwill/raftdemo/raft"
)

type Config struct {
	Raft   *raft.Config `json:"raft"`
	HTTP   *http.Config `json:"http"`
	Logger *log.Logger  `json:"logger"`
}

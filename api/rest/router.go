package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/musenwill/raftdemo/fsm"
	"io"
	"os"
)

type Ctx struct {
	Router *gin.Engine
}

func New(nodes []fsm.Prober) *Ctx {
	nodeMap := make(map[string]fsm.Prober)
	for _, n := range nodes {
		nodeMap[n.GetHost()] = n
	}

	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()

	nodeMgr := NodeMgr{nodes: nodeMap}
	(&NodeController{NodeMgr: &nodeMgr}).Register(router)

	return &Ctx{
		Router: router,
	}
}

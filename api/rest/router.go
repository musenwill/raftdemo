package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/musenwill/raftdemo/api"
	"github.com/musenwill/raftdemo/api/mgr"
	"io"
	"os"
)

type Ctx struct {
	Router *gin.Engine
	Ctx    *api.Context
}

func New(ctx *api.Context) *Ctx {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()

	nodeMgr := mgr.NodeMgr{Ctx: ctx}
	(&NodeController{NodeMgr: &nodeMgr}).Register(router)

	logMgr := mgr.LogMgr{Ctx: ctx}
	(&LogController{LogMgr: &logMgr}).Register(router)

	confMgr := mgr.ConfMgr{Ctx: ctx}
	(&ConfigController{ConfMgr: &confMgr}).Register(router)

	return &Ctx{
		Router: router,
		Ctx:    ctx,
	}
}

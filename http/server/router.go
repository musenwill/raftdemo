package server

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/http/mgr"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
)

func NewHandler(instances map[string]raft.NodeInstance, proxy proxy.Proxy, cfg *config.Config) *gin.Engine {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()

	ctx := &mgr.Context{
		Instances: instances,
		Proxy:     proxy,
		Cfg:       cfg,
	}

	nodeMgr := &mgr.NodeMgr{Ctx: ctx}
	(&NodeServer{NodeMgr: nodeMgr}).Register(router)

	entryMgr := &mgr.EntryMgr{Ctx: ctx}
	(&EntryServer{EntryMgr: entryMgr}).Register(router)

	cfgMgr := &mgr.CfgMgr{Ctx: ctx}
	(&ConfigServer{CfgMgr: cfgMgr}).Register(router)

	proxyMgr := &mgr.ProxyMgr{Ctx: ctx}
	(&ProxyServer{ProxyMgr: proxyMgr}).Register(router)

	ctx.NodeMgr = nodeMgr
	ctx.EntryMgr = entryMgr
	ctx.CfgMgr = cfgMgr
	ctx.ProxyMgr = proxyMgr

	router.GET("/ping", Ping)

	return router
}

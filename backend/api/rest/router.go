package rest

import (
	"io"
	"os"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gin-gonic/gin"
	"github.com/musenwill/raftdemo/api/mgr"
	"github.com/musenwill/raftdemo/api/types"
)

type Ctx struct {
	Router *gin.Engine
	Ctx    *types.Context
}

func New(ctx *types.Context) *Ctx {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()

	fsCss := assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "dist/css", Fallback: "index.html"}
	fsFonts := assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "dist/fonts", Fallback: "index.html"}
	fsImg := assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "dist/img", Fallback: "index.html"}
	fsJs := assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "dist/js", Fallback: "index.html"}
	fs := assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "dist", Fallback: "index.html"}

	router.StaticFS("/css", &fsCss)
	router.StaticFS("/fonts", &fsFonts)
	router.StaticFS("/img", &fsImg)
	router.StaticFS("/js", &fsJs)
	router.StaticFS("/favicon.ico", &fs)
	router.GET("/", func(c *gin.Context) {
		c.Writer.WriteHeader(200)
		indexHtml, _ := Asset("dist/index.html")
		_, _ = c.Writer.Write(indexHtml)
		c.Writer.Header().Add("Accept", "text/html")
		c.Writer.Flush()
	})

	nodeMgr := &mgr.NodeMgr{Ctx: ctx}
	(&NodeController{NodeMgr: nodeMgr}).Register(router)

	logMgr := &mgr.LogMgr{Ctx: ctx}
	(&LogController{LogMgr: logMgr}).Register(router)

	confMgr := &mgr.ConfMgr{Ctx: ctx}
	(&ConfigController{ConfMgr: confMgr}).Register(router)

	ctx.NodeMgr = nodeMgr
	ctx.LogMgr = logMgr
	ctx.ConfMgr = confMgr

	return &Ctx{
		Router: router,
		Ctx:    ctx,
	}
}

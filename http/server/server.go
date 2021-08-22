package server

import (
	"github.com/gin-gonic/gin"
	"github.com/musenwill/raftdemo/common"
	error2 "github.com/musenwill/raftdemo/http/error"
	"github.com/musenwill/raftdemo/http/mgr"
	"github.com/musenwill/raftdemo/http/types"
	"github.com/musenwill/raftdemo/raft"
)

type NodeServer struct {
	NodeMgr *mgr.NodeMgr
}

func (s *NodeServer) List(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	result, httpErr := s.NodeMgr.List()
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (m *NodeServer) Get(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param struct {
		Node string `uri:"node" binding:"required" json:"node"`
	}
	err := ctx.BindUri(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	var node raft.NodeInstance
	if param.Node == "leader" {
		node, httpErr = m.NodeMgr.GetLeader()
		if httpErr != nil {
			return
		}
	} else {
		node, httpErr = m.NodeMgr.GetNode(param.Node)
		if httpErr != nil {
			return
		}
	}

	ctx.JSON(200, m.NodeMgr.WrapNode(node))
}

func (m *NodeServer) Update(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var node struct {
		Node string `uri:"node" binding:"required" json:"node"`
	}
	err := ctx.BindUri(&node)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	var param struct {
		State string `json:"state"`
	}
	err = ctx.ShouldBind(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := m.NodeMgr.Update(node.Node, param.State)
	if httpErr != nil {
		return
	}
	ctx.JSON(200, result)
}

func (m *NodeServer) Register(router *gin.Engine) {
	g := router.Group("/nodes")
	g.GET("", m.List)
	g.GET(":node", m.Get)
	g.PUT(":node", m.Update)
}

type EntryServer struct {
	EntryMgr *mgr.EntryMgr
}

func (s *EntryServer) List(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param struct {
		Node string `uri:"node" binding:"required" json:"node"`
	}
	err := ctx.BindUri(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := s.EntryMgr.List(param.Node)
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (s *EntryServer) Get(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param struct {
		Node  string `uri:"node" binding:"required" json:"node"`
		Index int64  `uri:"index" json:"index"`
	}

	err := ctx.BindUri(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := s.EntryMgr.Get(param.Node, param.Index)
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (s *EntryServer) Add(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param struct {
		Data string `json:"data"`
	}
	err := ctx.ShouldBind(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := s.EntryMgr.Add([]byte(param.Data))
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (m *EntryServer) Register(router *gin.Engine) {
	g := router.Group("/")
	g.GET("/nodes/:node/logs", m.List)
	g.GET("/nodes/:node/logs/:index", m.Get)
	g.POST("logs", m.Add)
}

type ConfigServer struct {
	CfgMgr *mgr.CfgMgr
}

func (s *ConfigServer) Get(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	result, httpErr := s.CfgMgr.Get()
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (s *ConfigServer) Update(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param struct {
		Level string `json:"level"`
	}
	err := ctx.ShouldBind(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	httpErr = s.CfgMgr.Update(param.Level)
	if httpErr != nil {
		return
	}

	ctx.JSON(200, nil)
}

func (s *ConfigServer) Register(router *gin.Engine) {
	g := router.Group("/config")
	g.GET("", s.Get)
	g.PUT("", s.Update)
}

func Ping(ctx *gin.Context) {
	var httpErr *error2.HttpError
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	result := types.Ping{
		AppName:   common.AppName,
		Version:   common.Version,
		BuildTime: common.BuildTime,
		UpTime:    common.UpTime,
	}

	ctx.JSON(200, result)
}

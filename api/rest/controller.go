package rest

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/musenwill/raftdemo/api"
	error2 "github.com/musenwill/raftdemo/api/error"
	"github.com/musenwill/raftdemo/api/mgr"
)

type NodeController struct {
	NodeMgr *mgr.NodeMgr
}

func (p *NodeController) List(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	result, httpErr := p.NodeMgr.List()
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *NodeController) Add(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param AddNodeForm
	err := ctx.ShouldBind(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	ctx.JSON(200, param)
}

func (p *NodeController) Get(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param struct {
		Host string `uri:"host" binding:"required" json:"host"`
	}
	err := ctx.BindUri(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := p.NodeMgr.Get(param.Host)
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *NodeController) Update(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var tmp struct {
		Host string `uri:"host" binding:"required" json:"host"`
	}
	err := ctx.BindUri(&tmp)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}
	host := tmp.Host

	var param UpdateNodeForm
	err = ctx.ShouldBind(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := p.NodeMgr.Update(host, param.Timeout, param.Sleep)
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *NodeController) Delete(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	httpErr = error2.UnimplementedError(errors.New("update node has not implemented yet"))
}

func (p *NodeController) Register(router *gin.Engine) {
	g := router.Group("/v1")
	g.GET("/nodes", p.List)
	g.POST("/nodes", p.Add)
	g.GET("/nodes/:host", p.Get)
	g.PUT("/nodes/:host", p.Update)
	g.DELETE("/nodes/:host", p.Delete)
}

type LogController struct {
	LogMgr *mgr.LogMgr
}

func (p *LogController) List(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	result, httpErr := p.LogMgr.List()
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *LogController) Get(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param struct {
		Index *int64 `uri:"index" json:"index"`
	}
	err := ctx.BindUri(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := p.LogMgr.Get(*param.Index)
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *LogController) Add(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param AddLogForm
	err := ctx.ShouldBind(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	result, httpErr := p.LogMgr.Add(param.RequestID, param.Command)
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *LogController) Register(router *gin.Engine) {
	g := router.Group("/v1")
	g.GET("/logs", p.List)
	g.POST("logs", p.Add)
	g.GET("/logs/:index", p.Get)
}

type ConfigController struct {
	ConfMgr *mgr.ConfMgr
}

func (p *ConfigController) Get(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	result, httpErr := p.ConfMgr.Get()
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *ConfigController) Update(ctx *gin.Context) {
	var httpErr *error2.HttpError = nil
	defer func() {
		if httpErr != nil {
			ctx.JSON(httpErr.GetStatusCode(), httpErr)
		}
	}()

	var param api.ConfigInfo
	err := ctx.ShouldBind(&param)
	if err != nil {
		httpErr = error2.ParamError(err)
		return
	}

	httpErr = p.ConfMgr.Set(param)
	if httpErr != nil {
		return
	}

	result, httpErr := p.ConfMgr.Get()
	if httpErr != nil {
		return
	}

	ctx.JSON(200, result)
}

func (p *ConfigController) Register(router *gin.Engine) {
	g := router.Group("/v1")
	g.GET("/config", p.Get)
	g.PUT("/config", p.Update)
}

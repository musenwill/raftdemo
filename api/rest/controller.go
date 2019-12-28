package rest

import (
	"errors"
	"github.com/gin-gonic/gin"
	error2 "github.com/musenwill/raftdemo/api/error"
)

type NodeController struct {
	NodeMgr *NodeMgr
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

	httpErr = error2.UnimplementedError(errors.New("update node has not implemented yet"))
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
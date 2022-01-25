package mgr

import (
	"sort"

	error2 "github.com/musenwill/raftdemo/http/error"
	"github.com/musenwill/raftdemo/http/types"
	"github.com/musenwill/raftdemo/model"
)

type ProxyMgr struct {
	Ctx *Context
}

func (m *ProxyMgr) Get() (types.ListPipesResponse, *error2.HttpError) {
	pipes := m.Ctx.Proxy.BrokenPipes()
	sort.Sort(types.PipeList(pipes))
	return types.ListPipesResponse{
		Total:   len(pipes),
		Entries: pipes,
	}, nil
}

func (m *ProxyMgr) Update(pipes []model.Pipe) *error2.HttpError {
	err := m.Ctx.Proxy.SetPipeStates(pipes)
	if err != nil {
		return error2.ServerError(err)
	}
	return nil
}

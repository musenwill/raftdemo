package mgr

import (
	"github.com/musenwill/raftdemo/api"
	"github.com/musenwill/raftdemo/api/error"
	"github.com/musenwill/raftdemo/common"
)

type ConfMgr struct {
	Ctx *api.Context
}

func (p *ConfMgr) Get() (api.ConfigInfo, *error.HttpError) {
	nodes := p.Ctx.Conf.GetNodes()
	var nLst []string
	for _, v := range nodes {
		nLst = append(nLst, v.ID)
	}

	return api.ConfigInfo{
		LogLevel:          string(p.Ctx.Logger.GetLevel()),
		Nodes:             nLst,
		ReplicateTimeout:  p.Ctx.Conf.GetReplicateTimeout(),
		ReplicateUnitSize: p.Ctx.Conf.GetReplicateUnitSize(),
		MaxLogSize:        p.Ctx.Conf.GetMaxLogSize(),
	}, nil
}

func (p *ConfMgr) Set(cInfo api.ConfigInfo) *error.HttpError {
	if len(cInfo.LogLevel) > 0 {
		err := p.Ctx.Logger.SetLevel(common.LogLevel(cInfo.LogLevel))
		if err != nil {
			return error.ServerError(err)
		}
	}
	if cInfo.ReplicateTimeout > 0 {
		p.Ctx.Conf.SetReplicateTimeout(cInfo.ReplicateTimeout)
	}
	if cInfo.ReplicateUnitSize > 0 {
		p.Ctx.Conf.SetReplicateUnitSize(cInfo.ReplicateUnitSize)
	}
	if cInfo.MaxLogSize > 0 {
		p.Ctx.Conf.SetMaxLogSize(cInfo.MaxLogSize)
	}

	return nil
}

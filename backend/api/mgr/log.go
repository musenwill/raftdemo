package mgr

import (
	error2 "github.com/musenwill/raftdemo/api/error"
	"github.com/musenwill/raftdemo/api/types"
	"github.com/musenwill/raftdemo/model"
	"sort"
)

type LogMgr struct {
	Ctx *types.Context
}

func (p *LogMgr) List() (types.ListResponse, *error2.HttpError) {

	leader, httpErr := p.Ctx.NodeMgr.(*NodeMgr).GetLeader()
	if httpErr != nil {
		return types.ListResponse{}, httpErr
	}

	logs := leader.GetLogs()
	sort.Sort(model.LogList(logs))

	var result types.ListResponse
	for _, l := range logs {
		result.Entries = append(result.Entries, l)
	}
	result.Total = len(logs)
	return result, nil
}

func (p *LogMgr) Get(index int64) (model.Log, *error2.HttpError) {
	leader, httpErr := p.Ctx.NodeMgr.(*NodeMgr).GetLeader()
	if httpErr != nil {
		return model.Log{}, httpErr
	}

	log, err := leader.GetLog(index)
	if err != nil {
		return model.Log{}, error2.DataNotFoundError(err)
	}

	return log, nil
}

func (p *LogMgr) Add(requestID, command string) (model.Log, *error2.HttpError) {
	leader, httpErr := p.Ctx.NodeMgr.(*NodeMgr).GetLeader()
	if httpErr != nil {
		return model.Log{}, httpErr
	}

	logs, err := leader.AddLogs(model.Log{
		RequestID: requestID,
		Command:   model.StrCommand(command),
	})
	if err != nil {
		return model.Log{}, error2.ServerError(err)
	}

	return logs[0], nil
}

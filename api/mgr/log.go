package mgr

import (
	"errors"
	"github.com/musenwill/raftdemo/api"
	error2 "github.com/musenwill/raftdemo/api/error"
	"github.com/musenwill/raftdemo/fsm"
	"github.com/musenwill/raftdemo/model"
)

type LogMgr struct {
	Nodes map[string]fsm.Prober
}

func (p *LogMgr) List() (api.ListResponse, *error2.HttpError) {
	leader, httpErr := p.getLeader()
	if httpErr != nil {
		return api.ListResponse{}, httpErr
	}

	logs := leader.GetLogs()
	var result api.ListResponse
	for _, l := range logs {
		result.Entries = append(result.Entries, l)
	}
	result.Total = len(logs)
	return result, nil
}

func (p *LogMgr) Get(index int64) (model.Log, *error2.HttpError) {
	leader, httpErr := p.getLeader()
	if httpErr != nil {
		return model.Log{}, httpErr
	}

	log, err := leader.GetLog(index)
	if err != nil {
		return model.Log{}, error2.DataNotFoundError(err)
	}

	return log, nil
}

func (p *LogMgr) Add(log api.AddLogForm) (model.Log, *error2.HttpError) {
	leader, httpErr := p.getLeader()
	if httpErr != nil {
		return model.Log{}, httpErr
	}

	logs, err := leader.AddLogs(model.Log{
		RequestID: log.RequestID,
		Command:   log.Command,
	})
	if err != nil {
		return model.Log{}, error2.ServerError(err)
	}

	return logs[0], nil
}

func (p *LogMgr) getLeader() (fsm.Prober, *error2.HttpError) {
	var leader fsm.Prober

	for _, n := range p.Nodes {
		if n.GetState() == fsm.StateEnum.Leader {
			leader = n
			break
		}
	}

	if leader == nil {
		return nil, error2.ServerError(errors.New("leader undetermined"))
	}

	return leader, nil
}

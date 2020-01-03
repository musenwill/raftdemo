package mgr

import (
	"errors"
	"fmt"
	"github.com/musenwill/raftdemo/api"
	error2 "github.com/musenwill/raftdemo/api/error"
)

type NodeMgr struct {
	Ctx *api.Context
}

func (p *NodeMgr) List() (api.ListResponse, *error2.HttpError) {
	result := api.ListResponse{}
	result.Total = len(p.Ctx.NodeMap)

	for _, n := range p.Ctx.NodeMap {
		result.Entries = append(result.Entries, api.Node{
			Host:          n.GetHost(),
			Term:          n.GetTerm(),
			State:         string(n.GetState()),
			CommitIndex:   n.GetCommitIndex(),
			LastAppliedID: n.GetLastAppliedIndex(),
			Leader:        n.GetLeader(),
			VoteFor:       n.GetVoteFor(),
			Logs:          n.GetLogs(),
		})
	}

	return result, nil
}

func (p *NodeMgr) Get(host string) (api.Node, *error2.HttpError) {
	node, ok := p.Ctx.NodeMap[host]
	if !ok {
		return api.Node{}, error2.DataNotFoundError(errors.New(fmt.Sprintf("node %s not found", host)))
	}
	return api.Node{
		Host:          node.GetHost(),
		Term:          node.GetTerm(),
		State:         string(node.GetState()),
		CommitIndex:   node.GetCommitIndex(),
		LastAppliedID: node.GetLastAppliedIndex(),
		Leader:        "",
		VoteFor:       "",
		Logs:          node.GetLogs(),
	}, nil
}

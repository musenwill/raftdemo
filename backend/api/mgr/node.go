package mgr

import (
	"errors"
	"fmt"
	"github.com/musenwill/raftdemo/api"
	error2 "github.com/musenwill/raftdemo/api/error"
	"github.com/musenwill/raftdemo/fsm"
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
	node, httpErr := p.getNode(host)
	if httpErr != nil {
		return api.Node{}, httpErr
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

func (p *NodeMgr) Update(host string, timeout bool, sleep bool) (api.Node, *error2.HttpError) {
	node, httpErr := p.getNode(host)
	if httpErr != nil {
		return api.Node{}, httpErr
	}

	if timeout {
		node.ResetTimer()
	}

	if sleep {
		// at least sleep 2 rounds can leader loose its power
		node.Sleep(2 * p.Ctx.Conf.GetReplicateTimeout())
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

func (p *NodeMgr) getNode(host string) (fsm.Prober, *error2.HttpError) {
	node, ok := p.Ctx.NodeMap[host]
	if !ok {
		return nil, error2.DataNotFoundError(errors.New(fmt.Sprintf("node %s not found", host)))
	}
	return node, nil
}

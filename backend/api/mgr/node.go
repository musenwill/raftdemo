package mgr

import (
	"errors"
	"fmt"
	"github.com/musenwill/raftdemo/api"
	error2 "github.com/musenwill/raftdemo/api/error"
	"github.com/musenwill/raftdemo/fsm"
	"sort"
)

type NodeMgr struct {
	Ctx *api.Context
}

func (p *NodeMgr) List() (api.ListResponse, *error2.HttpError) {
	result := api.ListResponse{}
	result.Total = len(p.Ctx.NodeMap)
	var nodeList api.NodeList

	for _, n := range p.Ctx.NodeMap {
		nodeList = append(nodeList, p.wrapApiNode(n))
	}
	sort.Sort(nodeList)
	for _, v := range nodeList {
		result.Entries = append(result.Entries, v)
	}

	return result, nil
}

func (p *NodeMgr) Get(host string) (api.Node, *error2.HttpError) {
	node, httpErr := p.getNode(host)
	if httpErr != nil {
		return api.Node{}, httpErr
	}

	return p.wrapApiNode(node), nil
}

func (p *NodeMgr) wrapApiNode(node fsm.Prober) api.Node {
	var nodeState string
	if node.GetState() == fsm.StateEnum.Dummy {
		nodeState = api.NodeStateEnum.Stop
	} else {
		nodeState = api.NodeStateEnum.Start
	}

	return api.Node{
		Host:          node.GetHost(),
		Term:          node.GetTerm(),
		State:         string(node.GetState()),
		CommitIndex:   node.GetCommitIndex(),
		LastAppliedID: node.GetLastAppliedIndex(),
		Leader:        node.GetLeader(),
		VoteFor:       node.GetVoteFor(),
		Logs:          node.GetLogs(),
		NodeState:     nodeState,
	}
}

func (p *NodeMgr) Update(host string, nodeState string) (api.Node, *error2.HttpError) {
	node, httpErr := p.getNode(host)
	if httpErr != nil {
		return api.Node{}, httpErr
	}

	if nodeState == api.NodeStateEnum.Stop {
		node.TransferState(fsm.StateEnum.Dummy)
	} else if nodeState == api.NodeStateEnum.Start {
		if node.GetState() == fsm.StateEnum.Dummy {
			node.TransferState(fsm.StateEnum.Follower)
		}
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

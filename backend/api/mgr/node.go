package mgr

import (
	"errors"
	"fmt"
	error2 "github.com/musenwill/raftdemo/api/error"
	"github.com/musenwill/raftdemo/api/types"
	"github.com/musenwill/raftdemo/fsm"
	"sort"
)

type NodeMgr struct {
	Ctx *types.Context
}

func (p *NodeMgr) List() (types.ListResponse, *error2.HttpError) {
	result := types.ListResponse{}
	result.Total = len(p.Ctx.NodeMap)
	var nodeList types.NodeList

	for _, n := range p.Ctx.NodeMap {
		nodeList = append(nodeList, p.wrapApiNode(n))
	}
	sort.Sort(nodeList)
	for _, v := range nodeList {
		result.Entries = append(result.Entries, v)
	}

	return result, nil
}

func (p *NodeMgr) Get(host string) (types.Node, *error2.HttpError) {
	node, httpErr := p.GetNode(host)
	if httpErr != nil {
		return types.Node{}, httpErr
	}

	return p.wrapApiNode(node), nil
}

func (p *NodeMgr) wrapApiNode(node fsm.Prober) types.Node {
	var nodeState string
	if node.GetState() == fsm.StateEnum.Dummy {
		nodeState = types.NodeStateEnum.Stop
	} else {
		nodeState = types.NodeStateEnum.Start
	}

	return types.Node{
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

func (p *NodeMgr) Update(host string, nodeState string) (types.Node, *error2.HttpError) {
	node, httpErr := p.GetNode(host)
	if httpErr != nil {
		return types.Node{}, httpErr
	}

	if nodeState == types.NodeStateEnum.Stop {
		node.TransferState(fsm.StateEnum.Dummy)
	} else if nodeState == types.NodeStateEnum.Start {
		if node.GetState() == fsm.StateEnum.Dummy {
			node.TransferState(fsm.StateEnum.Follower)
		}
	}

	return types.Node{
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

func (p *NodeMgr) GetNode(host string) (fsm.Prober, *error2.HttpError) {
	node, ok := p.Ctx.NodeMap[host]
	if !ok {
		return nil, error2.DataNotFoundError(errors.New(fmt.Sprintf("node %s not found", host)))
	}
	return node, nil
}

func (p *NodeMgr) GetLeader() (fsm.Prober, *error2.HttpError) {
	var leader fsm.Prober

	for _, n := range p.Ctx.NodeMap {
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

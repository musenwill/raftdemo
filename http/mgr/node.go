package mgr

import (
	"fmt"
	"sort"

	"github.com/musenwill/raftdemo/model"

	error2 "github.com/musenwill/raftdemo/http/error"
	"github.com/musenwill/raftdemo/http/types"
	"github.com/musenwill/raftdemo/raft"
)

type NodeMgr struct {
	Ctx *Context
}

func (m *NodeMgr) List() (types.ListNodesResponse, *error2.HttpError) {
	result := types.ListNodesResponse{}
	result.Total = len(m.Ctx.Instances)

	var nodeList types.NodeList
	for _, n := range m.Ctx.Instances {
		nodeList = append(nodeList, m.WrapNode(n))
	}
	sort.Sort(nodeList)
	result.Entries = nodeList

	return result, nil
}

func (m *NodeMgr) Get(nodeID string) (types.Node, *error2.HttpError) {
	node, httpErr := m.GetNode(nodeID)
	if httpErr != nil {
		return types.Node{}, httpErr
	}
	return m.WrapNode(node), nil
}

func (m *NodeMgr) WrapNode(node raft.NodeInstance) types.Node {
	lastEntry := node.GetLastEntry()
	return types.Node{
		ID:            node.GetNodeID(),
		Term:          node.GetTerm(),
		State:         node.GetState().String(),
		CommitID:      node.GetCommitIndex(),
		LastAppliedID: node.GetLastAppliedIndex(),
		LastLogID:     lastEntry.Id,
		LastLogTerm:   lastEntry.Term,
		Leader:        node.GetLeader(),
		VoteFor:       node.GetVoteFor(),
		Readable:      node.Readable(),
	}
}

func (m *NodeMgr) Update(nodeID string, state string) (types.Node, *error2.HttpError) {
	stateRole, err := model.MapStateRole(state)
	if err != nil {
		return types.Node{}, error2.ParamError(err)
	}

	node, httpErr := m.GetNode(nodeID)
	if httpErr != nil {
		return types.Node{}, httpErr
	}

	if err := node.SwitchStateTo(stateRole); err != nil {
		return types.Node{}, error2.ServerError(err)
	}

	return m.WrapNode(node), nil
}

func (m *NodeMgr) GetNode(nodeID string) (raft.NodeInstance, *error2.HttpError) {
	node, ok := m.Ctx.Instances[nodeID]
	if !ok {
		return nil, error2.DataNotFoundError(fmt.Errorf("node %s not found", nodeID))
	}
	return node, nil
}

func (m *NodeMgr) GetLeader() (raft.NodeInstance, *error2.HttpError) {
	var leaders []raft.NodeInstance

	for _, n := range m.Ctx.Instances {
		if n.GetState() == model.StateRole_leader {
			leaders = append(leaders, n)
		}
	}

	if len(leaders) == 0 {
		return nil, error2.ServerError(fmt.Errorf("no leader"))
	}

	maxTerm := leaders[0]
	for _, n := range leaders {
		if n.GetTerm() > maxTerm.GetTerm() {
			maxTerm = n
		}
	}

	return maxTerm, nil
}

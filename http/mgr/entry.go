package mgr

import (
	"strings"

	error2 "github.com/musenwill/raftdemo/http/error"
	"github.com/musenwill/raftdemo/http/types"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
)

type EntryMgr struct {
	Ctx *Context
}

func (m *EntryMgr) List(nodeID string) (types.ListEntriesResponse, *error2.HttpError) {
	var result types.ListEntriesResponse

	node, err := m.getNode(nodeID)
	if err != nil {
		return result, err
	}

	entries := node.GetEntries()
	result.Entries = entries
	result.Total = len(entries)

	return result, nil
}

func (m *EntryMgr) Get(nodeID string, index int64) (model.Entry, *error2.HttpError) {
	node, err := m.getNode(nodeID)
	if err != nil {
		return model.Entry{}, err
	}

	entry, e := node.GetEntry(index)
	if e != nil {
		return model.Entry{}, error2.DataNotFoundError(e)
	}

	return entry, nil
}

func (m *EntryMgr) Add(data []byte) *error2.HttpError {
	leader, httpErr := m.Ctx.NodeMgr.GetLeader()
	if httpErr != nil {
		return httpErr
	}

	err := leader.AppendData(data)
	if err != nil {
		return error2.ServerError(err)
	}

	return nil
}

func (m *EntryMgr) getNode(nodeID string) (raft.NodeInstance, *error2.HttpError) {
	if nodeID == "" || strings.ToUpper(nodeID) == "LEADER" {
		leader, err := m.Ctx.NodeMgr.GetLeader()
		if err != nil {
			return nil, err
		}
		return leader, nil
	} else {
		node, err := m.Ctx.NodeMgr.GetNode(nodeID)
		if err != nil {
			return nil, err
		}
		return node, nil
	}
}

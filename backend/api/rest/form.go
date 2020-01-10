package rest

import (
	"errors"
	"github.com/musenwill/raftdemo/api/types"
)

type AddNodeForm struct {
	Host string `json:"host"`
}

type AddLogForm struct {
	RequestID string `json:"request_id"`
	Command   string `json:"command"`
}

type UpdateNodeForm struct {
	NodeState string `json:"node_state"`
}

func (p *UpdateNodeForm) Valid() error {
	var filedCount int
	if len(p.NodeState) > 0 {
		filedCount++
	}
	if filedCount > 1 {
		return errors.New("can not specify one field once")
	}

	if err := p.ValidNodeState(); err != nil {
		return err
	}

	return nil
}

func (p *UpdateNodeForm) ValidNodeState() error {
	var values = []string{"", types.NodeStateEnum.Start, types.NodeStateEnum.Stop}
	for _, v := range values {
		if p.NodeState == v {
			return nil
		}
	}
	return errors.New("state param invalid, should be one of [start, stop]")
}

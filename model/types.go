package model

import (
	fmt "fmt"
	"strings"

	"github.com/musenwill/raftdemo/common"
)

func MapStateRole(state string) (StateRole, error) {
	state = strings.ToLower(state)
	s, ok := StateRole_value[state]
	if !ok {
		return StateRole_None, fmt.Errorf("unknown state %s", state)
	}
	return StateRole(s), nil
}

func (e *Entry) Header() []string {
	return []string{"version", "id", "term", "type", "payload"}
}

func (e *Entry) Row() []string {
	payload := e.Payload
	if len(payload) > 32 {
		payload = payload[:32]
	}
	return []string{fmt.Sprintf("%d", e.Version), fmt.Sprintf("%d", e.Id), fmt.Sprintf("%d", e.Term), EntryType_name[int32(e.Type)], fmt.Sprintf("%s", payload)}
}

func (e *Entry) Form() *common.Form {
	form := common.NewForm()
	form.SetTags([]common.Tag{{K: "name", V: "entry"}})
	form.SetHeader(e.Header())
	form.AddRow(e.Row())
	return form
}

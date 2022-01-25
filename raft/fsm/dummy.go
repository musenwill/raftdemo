package fsm

import "github.com/musenwill/raftdemo/model"

type Dummy struct{}

func (d *Dummy) Enter() {}

func (d *Dummy) Leave() {}

func (d *Dummy) OnAppendEntries(param model.AppendEntries) model.Response {
	return model.Response{}
}

func (d *Dummy) OnRequestVote(param model.RequestVote) model.Response {
	return model.Response{}
}

func (d *Dummy) OnTimeout() {}

func (d *Dummy) State() model.StateRole {
	return model.StateRole_dummy
}

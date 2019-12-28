package api

import "github.com/musenwill/raftdemo/model"

type Node struct {
	Host          string `json:"host"`
	Term          int64  `json:"term"`
	State         string `json:"state"`
	CommitIndex   int64  `json:"commit_index"`
	LastAppliedID int64  `json:"last_applied_id"`

	Leader  string `json:"leader"`
	VoteFor string `json:"vote_for"`

	Logs []model.Log `json:"logs"`
}

type ListResponse struct {
	Total   int           `json:"total"`
	Entries []interface{} `json:"entries"`
}

type AddLogForm struct {
	RequestID string `json:"request_id"`
	Command   string `json:"command"`
}

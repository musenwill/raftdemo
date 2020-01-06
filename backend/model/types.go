package model

type Node struct {
	ID string
}

type Log struct {
	RequestID string `json:"request_id"`
	Command   []byte `json:"command"`
	Term      int64  `json:"term"`
}

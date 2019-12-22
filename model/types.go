package model

type Node struct {
	ID string
}

type Log struct {
	RequestID string
	Command   string
	Term      int64
}

package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type LocalTime time.Time

func (p LocalTime) MarshalJSON() ([]byte, error) {
	var stamp = fmt.Sprintf("\"%s\"", time.Time(p).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

type Command interface {
	json.Marshaler
	GetContent() []byte
}

type StrCommand string

func (p StrCommand) MarshalJSON() ([]byte, error) {
	var tmp = fmt.Sprintf("\"%s\"", p)
	return []byte(tmp), nil
}

func (p StrCommand) GetContent() []byte {
	return []byte(p)
}

type Node struct {
	ID string
}

type Log struct {
	RequestID  string    `json:"request_id"`
	Command    Command   `json:"command"`
	Term       int64     `json:"term"`
	AppendTime LocalTime `json:"append_time"`
	ApplyTime  LocalTime `json:"apply_time"`
}

// order by append time desc
type LogList []Log

func (p LogList) Len() int {
	return len(p)
}

func (p LogList) Less(i, j int) bool {
	return !(time.Time(p[i].AppendTime).UnixNano() < time.Time(p[j].AppendTime).UnixNano())
}

func (p LogList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

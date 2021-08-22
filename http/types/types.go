package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/model"
)

type Ping struct {
	AppName   string    `json:"app_name"`
	Version   string    `json:"version"`
	Branch    string    `json:"branch"`
	Commit    string    `json:"commit"`
	BuildTime string    `json:"build_time"`
	UpTime    time.Time `json:"up_time"`
}

type Node struct {
	ID            string `json:"id"`
	Term          int64  `json:"term"`
	State         string `json:"state"`
	CommitID      int64  `json:"commit_id"`
	LastAppliedID int64  `json:"last_applied_id"`
	LastLogID     int64  `json:"last_log_id"`
	LastLogTerm   int64  `json:"last_log_term"`

	Leader  string `json:"leader"`
	VoteFor string `json:"vote_for"`
}

func (n *Node) Header() []string {
	return []string{"id", "term", "state", "commitID", "lastAppliedID", "lastLogID", "lastLogTerm", "leader", "voteFor"}
}

func (n *Node) Row() []string {
	return []string{n.ID, fmt.Sprintf("%d", n.Term), n.State, fmt.Sprintf("%d", n.CommitID),
		fmt.Sprintf("%d", n.LastAppliedID), fmt.Sprintf("%d", n.LastLogID), fmt.Sprintf("%d",
			n.LastLogTerm), n.Leader, n.VoteFor}
}

func (n *Node) Form() *common.Form {
	form := common.NewForm()
	form.SetTags([]common.Tag{{K: "name", V: "node"}})
	form.SetHeader(n.Header())
	form.AddRow(n.Row())
	return form
}

type ListNodesResponse struct {
	Total   int    `json:"total"`
	Entries []Node `json:"entries"`
}

type ListEntriesResponse struct {
	Total   int           `json:"total"`
	Entries []model.Entry `json:"entries"`
}

type NodeList []Node

func (p NodeList) Len() int {
	return len(p)
}

func (p NodeList) Less(i, j int) bool {
	return strings.Compare(p[i].ID, p[j].ID) < 0
}

func (p NodeList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

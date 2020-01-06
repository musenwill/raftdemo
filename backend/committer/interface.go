package committer

import "github.com/musenwill/raftdemo/model"

type Committer interface {
	Commit(log model.Log) error
}

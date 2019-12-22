package committer

import "github.com/musenwill/raftdemo/proxy"

type Committer interface {
	Commit(log proxy.Log) error
}


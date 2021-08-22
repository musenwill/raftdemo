package mgr

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/raft"
)

type Context struct {
	Instances map[string]raft.NodeInstance
	Cfg       *config.Config

	NodeMgr  *NodeMgr
	CfgMgr   *CfgMgr
	EntryMgr *EntryMgr
}

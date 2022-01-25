package mgr

import (
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
)

type Context struct {
	Instances map[string]raft.NodeInstance
	Proxy     proxy.Proxy
	Cfg       *config.Config

	NodeMgr  *NodeMgr
	CfgMgr   *CfgMgr
	EntryMgr *EntryMgr
	ProxyMgr *ProxyMgr
}

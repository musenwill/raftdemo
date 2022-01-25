package mgr

import (
	"fmt"
	"strings"

	"github.com/musenwill/raftdemo/config"
	error2 "github.com/musenwill/raftdemo/http/error"
	"github.com/musenwill/raftdemo/log"
)

type CfgMgr struct {
	Ctx *Context
}

func (m *CfgMgr) Get() (config.Config, *error2.HttpError) {
	return *m.Ctx.Cfg, nil
}

func (m *CfgMgr) Update(level string) *error2.HttpError {
	levelUP := strings.ToUpper(level)
	logLevel := log.LogLevel(levelUP)
	if !logLevel.Valid() {
		return error2.ParamError(fmt.Errorf("unsupported log level %s, require %v", level, log.LogLevelList))
	}

	err := m.Ctx.Cfg.Logger.SetLevel(logLevel)
	if err != nil {
		return error2.ServerError(err)
	}
	m.Ctx.Cfg.Raft.LogLevel = logLevel

	return nil
}

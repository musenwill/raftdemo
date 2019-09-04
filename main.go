package main

import (
	"github.com/musenwill/raftdemo/common"
)

func main() {
	logConf := common.DefaultZapConfig("stderr", common.AppName+".log")
	logger := common.NewLogger(logConf)
	logger = logger.With("node", "node-1")
	logger.Infow("failed to fetch URL",
		"url", "http://example.com",
		"attempt", 3,
		"backoff", 1,
	)
}

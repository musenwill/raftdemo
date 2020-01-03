package cmd

import (
	"fmt"
	"github.com/musenwill/raftdemo/common"
	"github.com/urfave/cli"
	"os"
	"time"
)

func New() *cli.App {
	app := cli.NewApp()
	app.ErrWriter = os.Stdout
	app.EnableBashCompletion = true
	app.Name = common.AppName
	app.Usage = "A raft distribution algorithm simulation demo"
	app.Version = common.Version
	app.Author = "musenwill"
	app.Email = "musenwill@qq.com"
	app.Copyright = fmt.Sprintf("Copyright © 2020 - %v musenwill. All Rights Reserved.", time.Now().Year())
	app.Description = description

	app.Flags = []cli.Flag{LogLevelFlag, LogFileFlag, NodeCountFlag, MaxLogSizeFlag, ReplicateSizeFlag,
		ReplicateTimeoutFlag, CommitterTypeFlag, RestHostFlag, RestPortFlag}
	app.Action = start

	return app
}

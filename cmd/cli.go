package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/musenwill/raftdemo/common"
	"github.com/urfave/cli"
)

func New() *cli.App {
	app := cli.NewApp()
	app.ErrWriter = os.Stdout
	app.EnableBashCompletion = true
	app.Name = "rafts"
	app.Usage = "A simple raft distribution algorithm simulation demo"
	app.Version = common.PrintVersion()
	app.Author = "musenwill"
	app.Email = "musenwill@qq.com"
	app.Copyright = fmt.Sprintf("Copyright © 2020 - %v musenwill. All Rights Reserved.", time.Now().Year())
	app.Description = description

	app.Flags = []cli.Flag{LogLevelFlag, LogFileFlag, HTTPHostFlag, HTTPPortFlag, PProfPortFlag, NodeCountFlag,
		MaxReplicaBatchFlag, MaxDataSizeFlag, ElectionTimeoutFlag, ReplicateTimeoutFlag, CampaignTimeoutFlag,
		ElectionRandomFlag}
	app.Action = start

	return app
}

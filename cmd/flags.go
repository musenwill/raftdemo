package cmd

import (
	"fmt"
	"github.com/musenwill/raftdemo/common"
	"github.com/urfave/cli"
)

var LogLevelFlag = cli.StringFlag{
	Name:     "log-level",
	Usage:    fmt.Sprintf("set log level, options are %v", common.LogLevelList),
	EnvVar:   "RAFT_LOG_LEVEL",
	Required: false,
	Value:    string(common.LogLevelEnum.Info),
}

var LogFileFlag = cli.StringFlag{
	Name:     "log-file",
	Usage:    fmt.Sprintf("file specified to write logs"),
	Required: false,
	Value:    "raft.log",
}

var NodeCountFlag = cli.IntFlag{
	Name:     "n",
	Usage:    "how many nodes to start up a raft cluster",
	Required: false,
	Value:    1,
}

var MaxLogSizeFlag = cli.Int64Flag{
	Name:     "log-size",
	Usage:    "max data of a log entry in bytes",
	EnvVar:   "RAFT_MAX_LOG_SIZE",
	Required: false,
	Value:    1024 * 1024,
}

var ReplicateSizeFlag = cli.Int64Flag{
	Name:     "rep-size",
	Usage:    "max number of logs replicate from leader to followers once a time",
	EnvVar:   "RAFT_REP_SIZE",
	Required: false,
	Value:    1024,
}

var ReplicateTimeoutFlag = cli.Int64Flag{
	Name:     "rep-timeout",
	Usage:    "timeout value in milliseconds of nodes to send append entries or enter a new round of election",
	EnvVar:   "RAFT_REP_SIZE",
	Required: false,
	Value:    200,
}

var CommitterTypeFlag = cli.StringFlag{
	Name:     "committer",
	Usage:    "choose what kind of committer, options are [file-committer]",
	Required: false,
	Value:    "file-committer",
}

var RestHostFlag = cli.StringFlag{
	Name:     "host",
	Usage:    "host address to run restful backend server",
	Required: false,
	Value:    "127.0.0.1",
}

var RestPortFlag = cli.IntFlag{
	Name:     "port, p",
	Usage:    "port to run restful backend server",
	Required: false,
	Value:    8080,
}

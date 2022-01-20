package cmd

import (
	"fmt"
	"time"

	"github.com/musenwill/raftdemo/log"
	"github.com/urfave/cli"
)

var LogLevelFlag = cli.StringFlag{
	Name:     "log-level",
	Usage:    fmt.Sprintf("set log level, options are %v", log.LogLevelList),
	EnvVar:   "RAFT_LOG_LEVEL",
	Required: false,
	Value:    string(log.LogLevelEnum.Info),
}

var LogFileFlag = cli.StringFlag{
	Name:     "log-file",
	Usage:    "file to write log",
	Required: false,
	Value:    "raft.log",
}

var HTTPHostFlag = cli.StringFlag{
	Name:     "host",
	Usage:    "http server host",
	Required: false,
	Value:    "localhost",
}

var HTTPPortFlag = cli.IntFlag{
	Name:     "port",
	Usage:    "http server port",
	Required: false,
	Value:    8086,
}

var PProfPortFlag = cli.IntFlag{
	Name:     "pprof-port",
	Usage:    "pprof port",
	Required: false,
	Value:    6060,
}

var NodeCountFlag = cli.UintFlag{
	Name:     "node",
	Usage:    "how many nodes to start up a raft cluster",
	Required: false,
	Value:    1,
}

var MaxReplicaBatchFlag = cli.UintFlag{
	Name:     "replica-batch",
	Usage:    "at most number of logs can be sent to follower at once",
	Required: false,
	Value:    1000,
}

var MaxDataSizeFlag = cli.UintFlag{
	Name:     "data-size",
	Usage:    "at most data size a client can request to server",
	Required: false,
	Value:    1024 * 1024,
}

var ElectionTimeoutFlag = cli.DurationFlag{
	Name:     "election-timeout",
	Usage:    "election timeout in milliseconds",
	Required: false,
	Value:    2000 * time.Millisecond,
}

var ReplicateTimeoutFlag = cli.DurationFlag{
	Name:     "replicate-timeout",
	Usage:    "replicate timeout in milliseconds to push append entries",
	Required: false,
	Value:    500 * time.Millisecond,
}

var CampaignTimeoutFlag = cli.DurationFlag{
	Name:     "campaign-timeout",
	Usage:    "campaign timeout in milliseconds for candidate to start a election in new term",
	Required: false,
	Value:    1000 * time.Millisecond,
}

var ElectionRandomFlag = cli.DurationFlag{
	Name:     "election-random",
	Usage:    "random time in milliseconds for candidate to wait before election",
	Required: false,
	Value:    500 * time.Millisecond,
}

var RequestTimeoutFlag = cli.DurationFlag{
	Name:     "request-timeout",
	Usage:    "rpc request timeout in milliseconds",
	Required: false,
	Value:    2 * time.Second,
}

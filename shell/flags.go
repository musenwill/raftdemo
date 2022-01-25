package shell

import (
	"github.com/urfave/cli"
)

var HTTPHostFlag = cli.StringFlag{
	Name:     "host",
	Usage:    "raft server host",
	Required: false,
	Value:    "localhost",
}

var HTTPPortFlag = cli.IntFlag{
	Name:     "port",
	Usage:    "raft server port",
	Required: false,
	Value:    8086,
}

var ExecuteFlag = cli.StringFlag{
	Name:     "exec",
	Usage:    "exec command",
	Required: false,
}

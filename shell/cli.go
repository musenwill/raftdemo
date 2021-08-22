package shell

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/musenwill/raftdemo/common"
	"github.com/peterh/liner"
	"github.com/urfave/cli"
)

func New() *cli.App {
	app := cli.NewApp()
	app.ErrWriter = os.Stdout
	app.EnableBashCompletion = true
	app.Name = common.AppName
	app.Usage = "Raft demo client"
	app.Version = common.Version
	app.Author = "musenwill"
	app.Email = "musenwill@qq.com"
	app.Copyright = fmt.Sprintf("Copyright © 2020 - %v musenwill. All Rights Reserved.", time.Now().Year())

	app.Flags = []cli.Flag{HTTPHostFlag, HTTPPortFlag, ExecuteFlag}
	app.Action = start

	return app
}

func start(c *cli.Context) error {
	host := c.String(HTTPHostFlag.Name)
	port := c.Uint(HTTPPortFlag.Name)

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	url := url.URL{
		Scheme: "http",
		Host:   addr,
	}
	app := &App{
		url:        url,
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}
	if err := app.Open(); err != nil {
		return err
	}
	defer app.Close()

	if err := app.Ping(nil); err != nil {
		return err
	}

	OsSignals := make(chan os.Signal)
	signal.Notify(OsSignals, syscall.SIGINT, syscall.SIGTERM)

	if c.IsSet(ExecuteFlag.Name) {
		exec := c.String(ExecuteFlag.Name)
		stmts, err := ParseStatements(exec)
		if err != nil {
			return err
		}

		for _, s := range stmts {
			select {
			case <-OsSignals:
				return nil
			default:
			}
			if err := app.Exec(s); err != nil {
				return err
			}
		}
	} else {
		liner := liner.NewLiner()
		defer liner.Close()

		for {
			select {
			case <-OsSignals:
				return nil
			default:
			}

			cmd, err := liner.Prompt("> ")
			if err != nil {
				os.Stderr.WriteString(err.Error() + "\n")
				continue
			}
			stmt, err := ParseStatement(cmd)
			if err != nil {
				os.Stderr.WriteString(err.Error() + "\n")
				continue
			}
			if stmt == nil {
				continue
			}
			liner.AppendHistory(stmt.String())

			if stmt.GetName() == EXIT {
				return nil
			}

			if err := app.Exec(stmt); err != nil {
				os.Stderr.WriteString(err.Error() + "\n")
			}
		}
	}

	return nil
}

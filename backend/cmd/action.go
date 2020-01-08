package cmd

import (
	"context"
	"fmt"
	"github.com/musenwill/raftdemo/api"
	"github.com/musenwill/raftdemo/api/rest"
	"github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/fsm"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/urfave/cli"
	"log"
	"net/http"
)

func start(c *cli.Context) error {
	logLevel := c.String("log-level")
	logFile := c.String("log-file")
	nodeNumber := c.Int("n")
	maxLogSize := c.Int64("log-size")
	replicateSize := c.Int64("rep-size")
	replicateTimeout := c.Int64("rep-timeout")
	host := c.String("host")
	port := c.Int("port")

	var nodes []model.Node
	for i := 1; i <= nodeNumber; i++ {
		nodes = append(nodes, model.Node{ID: fmt.Sprintf("node-%02d", i)})
	}

	logger, err := initLogger(common.LogLevel(logLevel), logFile)
	if err != nil {
		return err
	}
	conf, err := initConfig(nodes, replicateSize, replicateTimeout, maxLogSize)
	if err != nil {
		return err
	}
	chanProxy, err := initProxy(nodes)
	if err != nil {
		return err
	}
	probers, err := startRaft(chanProxy, conf, logger)
	if err != nil {
		return err
	}

	nodeMap := make(map[string]fsm.Prober)
	for _, v := range probers {
		nodeMap[v.GetHost()] = v
	}

	ctx := &api.Context{
		NodeMap: nodeMap,
		Proxy:   chanProxy,
		Conf:    conf,
		Logger:  logger,
	}

	srv, err := startRest(host, port, ctx)
	if err != nil {
		return err
	}

	handleExit(probers, srv)

	return nil
}

func initLogger(level common.LogLevel, logFile string) (*common.Logger, error) {
	return common.NewLogger(level, logFile)
}

func initConfig(nodes []model.Node, replicateUnitSize int64, replicateTimeout int64, maxLogSize int64) (config.Config, error) {
	conf := config.NewDefaultConfig(nodes, replicateUnitSize, replicateTimeout, maxLogSize)
	return conf, conf.Check()
}

func initProxy(nodes []model.Node) (*proxy.ChanProxy, error) {
	chanProxy := proxy.NewChanProxy()
	chanProxy.Register(nodes)
	return chanProxy, nil
}

func startRaft(chanProxy *proxy.ChanProxy, conf config.Config, logger *common.Logger) ([]fsm.Prober, error) {
	var servers []fsm.Prober
	for _, node := range conf.GetNodes() {
		cmt, err := committer.NewFileCommitter(fmt.Sprintf("commit-%s.txt", node.ID))
		if err != nil {
			return nil, err
		}
		server := fsm.NewServer(node.ID, cmt, chanProxy, conf, logger)
		server.Start()
		servers = append(servers, server)
	}
	return servers, nil
}

func startRest(host string, port int, ctx *api.Context) (*http.Server, error) {
	restSrv := rest.New(ctx)
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: restSrv.Router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv, nil
}

func handleExit(probers []fsm.Prober, srv *http.Server) {
	common.GracefulExit(func() error {
		for _, p := range probers {
			p.Stop()
		}
		return srv.Shutdown(context.Background())
	})
}

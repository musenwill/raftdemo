package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/musenwill/raftdemo/http/server"
	"github.com/musenwill/raftdemo/log"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
	"github.com/musenwill/raftdemo/raft/fsm"

	"github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/config"
	http2 "github.com/musenwill/raftdemo/http"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/urfave/cli"
)

func start(c *cli.Context) error {
	cfg, err := initConfig(c)
	if err != nil {
		return err
	}

	var nodes []raft.Node
	for i := 1; i <= int(cfg.Raft.Nodes); i++ {
		nodes = append(nodes, raft.Node{ID: fmt.Sprintf("node%02d", i)})
	}

	proxy := proxy.NewChanProxy(nodes, cfg.Raft.RequestTimeout)

	instances, err := startRaft(nodes, proxy, cfg.Raft)
	if err != nil {
		return err
	}
	srv, err := startHTTP(instances, cfg)
	if err != nil {
		handleExit(true, instances, nil)
		return err
	}

	startPprof(cfg.HTTP.PprofPort)
	handleExit(false, instances, srv)

	return nil
}

func initProxy(nodes []raft.Node, requestTimeout time.Duration) (*proxy.ChanProxy, error) {
	chanProxy := proxy.NewChanProxy(nodes, requestTimeout)
	return chanProxy, nil
}

func startRaft(nodes []raft.Node, proxy proxy.Proxy, cfg *raft.Config) (instances map[string]raft.NodeInstance, err error) {
	instances = make(map[string]raft.NodeInstance)
	defer func() {
		if err != nil {
			for _, v := range instances {
				_ = v.Close()
			}
		}
	}()

	for i, node := range nodes {
		cmt := committer.NewImmCommitter()
		instance := fsm.NewInstance(node.ID, cmt, proxy, cfg)
		if i == 0 {
			instance.State = model.StateRole_Leader
		}
		if err := instance.Open(); err != nil {
			return instances, err
		}
		instances[node.ID] = instance
	}
	return instances, nil
}

func startHTTP(instaces map[string]raft.NodeInstance, cfg *config.Config) (*http.Server, error) {
	handler := server.NewHandler(instaces, cfg)
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HTTP.HttpHost, cfg.HTTP.HttpPort),
		Handler: handler,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	return srv, nil
}

func startPprof(port uint) {
	go func() {
		fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()
}

func handleExit(rightNow bool, instances map[string]raft.NodeInstance, srv *http.Server) {
	common.GracefulExit(rightNow, func() error {
		if instances != nil {
			for _, s := range instances {
				if err := s.Close(); err != nil {
					fmt.Printf("close %s: %s\n", s.GetNodeID(), err)
				}
			}
		}
		if srv != nil {
			return srv.Shutdown(context.Background())
		}
		return nil
	})
}

func initConfig(c *cli.Context) (*config.Config, error) {
	cfg := &config.Config{
		Raft: &raft.Config{},
		HTTP: &http2.Config{},
	}

	cfg.HTTP.HttpHost = c.String(HTTPHostFlag.Name)
	cfg.HTTP.HttpPort = c.Uint(HTTPPortFlag.Name)
	cfg.HTTP.PprofPort = c.Uint(PProfPortFlag.Name)

	cfg.Raft.Nodes = c.Uint(NodeCountFlag.Name)
	cfg.Raft.MaxDataSize = c.Uint(MaxDataSizeFlag.Name)
	cfg.Raft.ReplicateMaxBatch = c.Uint(MaxReplicaBatchFlag.Name)
	cfg.Raft.ReplicateTimeout = c.Duration(ReplicateTimeoutFlag.Name)
	cfg.Raft.ElectionTimeout = c.Duration(ElectionTimeoutFlag.Name)
	cfg.Raft.ElectionRandom = c.Duration(ElectionRandomFlag.Name)
	cfg.Raft.CampaignTimeout = c.Duration(CampaignTimeoutFlag.Name)
	cfg.Raft.RequestTimeout = c.Duration(RequestTimeoutFlag.Name)

	cfg.Raft.LogLevel = log.LogLevel(c.String(LogLevelFlag.Name))
	cfg.Raft.LogFile = c.String(LogFileFlag.Name)
	logger, err := log.NewLogger(cfg.Raft.LogLevel, cfg.Raft.LogFile)
	if err != nil {
		return nil, err
	}
	cfg.Raft.Logger = logger

	return cfg, nil
}

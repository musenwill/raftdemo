package main

import (
	"context"
	"fmt"
	"github.com/musenwill/raftdemo/api/rest"
	committer2 "github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/model"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/fsm"
	"github.com/musenwill/raftdemo/proxy"
)

func main() {
	time.Sleep(time.Second * 2)

	nodeNum, _ := strconv.Atoi(os.Args[1])
	var nodes []model.Node
	for i := 1; i <= nodeNum; i++ {
		nodes = append(nodes, model.Node{ID: fmt.Sprintf("node-%d", i)})
	}
	chanProxy := proxy.NewChanProxy()
	chanProxy.Register(nodes)

	logger := common.NewLogger(common.DefaultZapConfig("raft.log"))
	conf := config.NewDefaultConfig(nodes, 1000, 1000)
	conf.Check()
	committer, err := committer2.NewFileCommitter("commit.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	var servers []fsm.Prober
	for _, node := range nodes {
		server := fsm.NewServer(node.ID, committer, chanProxy, conf, logger)
		server.Start()
		servers = append(servers, server)
	}

	// start restful server
	restSrv := rest.New(servers)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: restSrv.Router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// graceful shutdown servers
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down restful Server ...")
	err = srv.Shutdown(context.Background())
	fmt.Println("Shutting down fsm Server ...", err)
	for _, srv := range servers {
		srv.Stop()
	}
}

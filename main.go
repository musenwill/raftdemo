package main

import (
	"fmt"
	committer2 "github.com/musenwill/raftdemo/committer"
	"github.com/musenwill/raftdemo/model"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
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
	conf := config.NewDefaultConfig(nodes, 1000, 1)
	conf.Check()
	committer, _ := committer2.NewFileCommitter("commit.txt")

	for _, node := range nodes {
		server := fsm.NewServer(node.ID, committer, chanProxy, conf, logger)
		server.Start()
	}

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

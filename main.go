package main

import (
	"fmt"
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
	var nodes []config.Node
	for i := 1; i <= nodeNum; i++ {
		nodes = append(nodes, config.Node{ID: fmt.Sprintf("node-%d", i)})
	}
	proxy.Config(nodes)

	logger := common.NewLogger(common.DefaultZapConfig("raft.log"))
	config := &config.Config{Timeout: 5000, MaxReplicate: 1, Nodes: nodes}
	committer, _ := fsm.NewFileCommitter("commit.txt")

	for _, node := range nodes {
		server := fsm.NewServer(node.ID, committer, config, logger)
		server.Run()
	}

	http.ListenAndServe(":8080", nil)
}

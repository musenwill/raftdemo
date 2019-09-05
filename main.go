package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/musenwill/raftdemo/common"
	"github.com/musenwill/raftdemo/config"
	"github.com/musenwill/raftdemo/fsm"
	"github.com/musenwill/raftdemo/proxy"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":80", nil))
	}()

	nodeNum, _ := strconv.Atoi(os.Args[1])
	var nodes []config.Node
	for i := 1; i <= nodeNum; i++ {
		nodes = append(nodes, config.Node{ID: fmt.Sprintf("node-%d", i)})
	}
	proxy.Config(nodes)

	logger := common.NewLogger(common.DefaultZapConfig("raft.log"))
	config := &config.Config{Timeout: 20, MaxReplicate: 1, Nodes: nodes}
	committer, _ := fsm.NewFileCommitter("commit.txt")

	for _, node := range nodes {
		server := fsm.NewServer(node.ID, committer, config, logger)
		server.Run()
	}

	stop := make(chan bool)
	<-stop
}

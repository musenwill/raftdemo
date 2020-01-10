package main

import (
	"fmt"
	"github.com/musenwill/raftdemo/cmd"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	fmt.Println(cmd.New().Run(os.Args))
}

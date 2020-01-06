package main

import (
	"fmt"
	"github.com/musenwill/raftdemo/cmd"
	_ "net/http/pprof"
	"os"
)

func main() {
	fmt.Println(cmd.New().Run(os.Args))
}

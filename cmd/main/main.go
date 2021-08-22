package main

import (
	"fmt"
	"os"
	"time"

	"github.com/musenwill/raftdemo/cmd"
	"github.com/musenwill/raftdemo/common"
)

func main() {
	common.UpTime = time.Now()
	err := cmd.New().Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

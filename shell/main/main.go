package main

import (
	"fmt"
	"os"

	"github.com/musenwill/raftdemo/shell"
)

func main() {
	err := shell.New().Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

package common

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GracefulExit(rightNow bool, f func() error) {
	do := func() {
		fmt.Println("Shutting down ...")
		err := f()
		if err != nil {
			fmt.Println("shutdown server, error: ", err)
		} else {
			fmt.Println("shutdown server success")
		}
	}

	if rightNow {
		do()
		return
	}

	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	do()
}

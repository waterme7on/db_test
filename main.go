package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/prestodb/presto-go-client/presto"
)

var tm = ThreadsPool(3)
var scaler = Scaler{
	podPrefix: "gourdstore-slave",
}

func main() {
	//创建监听退出chan
	c := make(chan os.Signal)
	quit := false
	//监听指定信号 ctrl+c kill
	ctx, cancel := context.WithCancel(context.TODO())
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	var wg sync.WaitGroup
	workerSize := 0
	wg.Add(2 + workerSize)
	for i := 0; i < workerSize; i++ {
		go func(i int) {
			w := Worker{}
			err := w.Init(DSN, i)
			defer w.Close()
			defer wg.Done()
			if err != nil {
				return
			}
			w.Run(ctx)
		}(i)
	}
	go func() {
		scaler.Run(ctx)
		defer wg.Done()
	}()
	go func() {
		tm.Run(ctx)
		defer wg.Done()
	}()

	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			fmt.Println("Main routine Exit...", s)
			cancel()
			quit = true
		default:
			fmt.Println("other signal", s)
		}
		if quit {
			break
		}
	}
	fmt.Println("Main routine Start exit...")
	fmt.Println("Execute clean and wait for subroutines to quit...")
	wg.Wait()
	fmt.Println("Main routine end exit...")
}

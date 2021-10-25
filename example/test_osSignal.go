package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func main() {
	//创建监听退出chan
	c := make(chan os.Signal)
	quit := false
	//监听指定信号 ctrl+c kill
	ctx, cancel := context.WithCancel(context.TODO())
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	wg := sync.WaitGroup{}
	fmt.Println("Before Routine cnt:", runtime.NumGoroutine())
	go func(ctx context.Context) {
		wg.Add(1)
		defer wg.Done()
		fmt.Println("Sub routine Start...")
		sum := 0
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Sub routine quit...")
				return
			default:
				sum++
				go func() {
					fmt.Println("sum:", sum)
				}()
			}
			time.Sleep(time.Second)
			fmt.Println("Routine cnt:", runtime.NumGoroutine())
		}
	}(ctx)
	fmt.Println("After Routine cnt:", runtime.NumGoroutine())
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			fmt.Println("Main routine Exit...", s)
			cancel()
			quit = true
		case syscall.SIGUSR1:
			fmt.Println("usr1 signal", s)
		case syscall.SIGUSR2:
			fmt.Println("usr2 signal", s)
		default:
			fmt.Println("other signal", s)
		}
		if quit {
			break
		}
	}
	fmt.Println("Start Exit...")
	fmt.Println("Execute Clean...")
	wg.Wait()
	fmt.Println("After Routine cnt:", runtime.NumGoroutine())
	fmt.Println("Main routine End Exit...")
}

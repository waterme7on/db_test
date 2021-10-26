package main

import (
	"context"
	"log"
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
	log.Println("Before Routine cnt:", runtime.NumGoroutine())
	go func(ctx context.Context) {
		wg.Add(1)
		defer wg.Done()
		log.Println("Sub routine Start...")
		sum := 0
		for {
			select {
			case <-ctx.Done():
				log.Println("Sub routine quit...")
				return
			default:
				sum++
				go func() {
					log.Println("sum:", sum)
				}()
			}
			time.Sleep(time.Second)
			log.Println("Routine cnt:", runtime.NumGoroutine())
		}
	}(ctx)
	log.Println("After Routine cnt:", runtime.NumGoroutine())
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Println("Main routine Exit...", s)
			cancel()
			quit = true
		case syscall.SIGUSR1:
			log.Println("usr1 signal", s)
		case syscall.SIGUSR2:
			log.Println("usr2 signal", s)
		default:
			log.Println("other signal", s)
		}
		if quit {
			break
		}
	}
	log.Println("Start Exit...")
	log.Println("Execute Clean...")
	wg.Wait()
	log.Println("After Routine cnt:", runtime.NumGoroutine())
	log.Println("Main routine End Exit...")
}

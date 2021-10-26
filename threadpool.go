package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"
)

type threadsPool struct {
	mu   *sync.Mutex
	cnt  int
	size int
	// cnt为已经申请的数量
	// size为可申请数量
	// cnt <= size
}

func ThreadsPool(size int) *threadsPool {
	return &threadsPool{
		mu:   &sync.Mutex{},
		cnt:  0,
		size: size,
	}
}

func (tm *threadsPool) Alloc() int {
	ret := -1
	tm.mu.Lock()
	if tm.cnt < tm.size {
		// cnt < size才能申请到
		tm.cnt++
		ret = tm.cnt
	}
	tm.mu.Unlock()
	return ret
}

func (tm *threadsPool) Free(key int) int {
	ret := -1
	tm.mu.Lock()
	if tm.cnt > 0 {
		tm.cnt--
		ret = 0
	}
	tm.mu.Unlock()
	return ret
}

func (tm *threadsPool) Resize(size int) bool {
	// 直接resize
	tm.size = size
	return true
}

func (tm *threadsPool) CollectResult(ctx context.Context, c chan string) {
	file, _ := os.Create(fmt.Sprintf("result-%v.csv", time.Now().Unix()))
	defer file.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case input := <-c:
			file.WriteString(input)
		default:
			continue
		}
	}
}

func (tm *threadsPool) Run(ctx context.Context) {
	file, _ := os.Create(fmt.Sprintf("workload-%v.csv", time.Now().Unix()))
	defer file.Close()
	// file, _ := os.OpenFile("workload.csv", os.O_CREATE|os.O_APPEND, 0777)
	resizeTimer := time.NewTimer(ResizeInterval)
	for {
		time.Sleep(MonitorInterval)
		select {
		case <-ctx.Done():
			return
		case <-resizeTimer.C:
			tm.mu.Lock()
			newSize := rand.NormFloat64()*10 + 5
			if false && newSize >= 0 && newSize <= 10 {
				// DEBUG don't resize
				log.Printf("%v", newSize)
				log.Printf("Running queries size change from %v to %v", tm.size, int(newSize))
				tm.size = int(newSize)
			}
			resizeTimer.Reset(ResizeInterval)
			log.Printf("Current runing query: %d/%d(go routines: %d)\n", tm.cnt, tm.size, runtime.NumGoroutine())
			outputString := fmt.Sprintf("%v, %d, %d\n", time.Now().Unix(), tm.cnt, tm.size)
			n, err := file.WriteString(outputString)
			log.Printf("ThreadsPool write to file: %d, %v\n", n, err)
			tm.mu.Unlock()
		default:
			tm.mu.Lock()
			log.Printf("Current runing query: %d/%d(go routines: %d)\n", tm.cnt, tm.size, runtime.NumGoroutine())
			outputString := fmt.Sprintf("%v, %d, %d\n", time.Now().Unix(), tm.cnt, tm.size)
			n, err := file.WriteString(outputString)
			log.Printf("ThreadsPool write to file: %d, %v\n", n, err)
			tm.mu.Unlock()
		}
	}
}
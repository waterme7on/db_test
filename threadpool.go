package main

/*
	threadsPool - 实现对查询的并发数量的控制
*/

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
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
	file, _ := os.Create(fmt.Sprintf("result-%v-%v-%v.csv", MaxQuerySize, DynamicWorkload, ScalerOn))
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
	file, _ := os.Create(fmt.Sprintf("workload-%v-%v-%v.csv", MaxQuerySize, DynamicWorkload, ScalerOn))

	defer file.Close()

	var wlfile *os.File
	var scanner *bufio.Scanner
	if WorkloadFile != "" {
		wlfile, _ = os.Open(WorkloadFile)
		log.Print(WorkloadFile, wlfile)
		scanner = bufio.NewScanner(wlfile)
	} else {
		wlfile = nil
	}
	resizeTimer := time.NewTimer(ResizeInterval)
	rand.Seed(time.Now().Unix())
	for {
		time.Sleep(MonitorInterval)
		select {
		case <-ctx.Done():
			return
		case <-resizeTimer.C:
			if !DynamicWorkload {
				continue
			}
			tm.mu.Lock()
			newSize := 0
			if wlfile == nil {
				newSize = int(rand.NormFloat64()*10 + 5)
			} else {
				ok := scanner.Scan()
				if !ok {
					wlfile.Seek(0, 0)
					scanner = bufio.NewScanner(wlfile)
					scanner.Scan()
				}
				newSize, _ = strconv.Atoi(scanner.Text())
				log.Printf("newsize %v|%v", scanner.Text(), newSize)
			}
			if newSize >= 0 && newSize <= MaxQuerySize {
				// DEBUG don't resize
				log.Printf("Running queries size change from %v to %v", tm.size, int(newSize))
				tm.size = int(newSize)
			}
			resizeTimer.Reset(ResizeInterval)
			log.Printf("Current runing query: %d/%d(go routines: %d)\n", tm.cnt, tm.size, runtime.NumGoroutine())
			outputString := fmt.Sprintf("%v, %d, %d\n", time.Now().Format("2006-01-02 15:04:05"), tm.cnt, tm.size)
			file.WriteString(outputString)
			// log.Printf("ThreadsPool write to file: %d, %v\n", n, err)
			tm.mu.Unlock()
		default:
			tm.mu.Lock()
			log.Printf("Current runing query: %d/%d(go routines: %d)\n", tm.cnt, tm.size, runtime.NumGoroutine())
			outputString := fmt.Sprintf("%v, %d, %d\n", time.Now().Format("2006-01-02 15:04:05"), tm.cnt, tm.size)
			file.WriteString(outputString)
			// log.Printf("ThreadsPool write to file: %d, %v\n", n, err)
			tm.mu.Unlock()
		}
	}
}

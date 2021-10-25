package main

import "sync"

type ThreadsManager struct {
	mu   *sync.Mutex
	cnt  int
	size int
	// cnt为已经申请的数量
	// size为可申请数量
	// cnt <= size
}

func (tm *ThreadsManager) Alloc() int {
	ret := -1
	tm.mu.Lock()
	if tm.size == 0 {
		return ret
	}
	if tm.cnt < tm.size {
		tm.cnt++
		ret = tm.cnt
	}
	tm.mu.Unlock()
	return ret
}

func (tm *ThreadsManager) Free(key int) int {
	ret := -1
	tm.mu.Lock()
	if key > tm.size {
		return ret
	}
	if tm.cnt > 0 {
		tm.cnt--
		ret = 0
	}
	tm.mu.Unlock()
	return ret
}

func (tm *ThreadsManager) Resize(size int) bool {
	tm.mu.Lock()
	if size < tm.cnt {
		return false
	}
	tm.cnt = size
	tm.mu.Unlock()
	return true
}

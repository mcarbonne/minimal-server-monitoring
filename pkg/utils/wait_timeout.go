package utils

import (
	"sync"
	"time"
)

func WaitGroupChan(wg *sync.WaitGroup) chan struct{} {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	return c
}

// return true on timeout
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	select {
	case <-WaitGroupChan(wg):
		return false
	case <-time.After(timeout):
		return true
	}
}

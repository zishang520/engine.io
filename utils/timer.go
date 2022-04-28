package utils

import (
	"sync"
	"time"
)

type Timer struct {
	t       chan struct{}
	isClose bool

	// mu guards hijackedv
	mu sync.RWMutex
}

func SetTimeOut(fn func(), sleep time.Duration) *Timer {
	timer := &Timer{
		t:       make(chan struct{}),
		isClose: false,
	}
	t := time.NewTimer(sleep)
	go func() {
		select {
		case <-t.C:
			go fn()
		case <-timer.t:
			t.Stop()
			return
		}
	}()
	return timer
}

func SetInterval(fn func(), sleep time.Duration) *Timer {
	timer := &Timer{
		t:       make(chan struct{}),
		isClose: false,
	}
	t := time.NewTimer(sleep)
	go func() {
		for {
			select {
			case <-t.C:
				go fn()
				t.Reset(sleep)
			case <-timer.t:
				t.Stop()
				return
			}
		}
	}()
	return timer
}

func ClearInterval(timer *Timer) {
	ClearTimeout(timer)
}

func ClearTimeout(timer *Timer) {
	if timer != nil {
		timer.mu.Lock()
		defer timer.mu.Unlock()

		if !timer.isClose {
			close(timer.t)
			timer.isClose = true
		}
	}
}

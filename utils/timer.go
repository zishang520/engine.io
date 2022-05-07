package utils

import (
	"sync"
	"time"
)

type Timer struct {
	t       chan struct{}
	timer   *time.Timer
	sleep   time.Duration
	isClose bool

	// mu guards hijackedv
	mu sync.Mutex
}

func (t *Timer) Refresh() *Timer {
	t.timer.Reset(t.sleep)
	return t
}

func SetTimeOut(fn func(), sleep time.Duration) *Timer {
	timeout := &Timer{
		t:       make(chan struct{}),
		timer:   time.NewTimer(sleep),
		sleep:   sleep,
		isClose: false,
	}
	go func() {
		select {
		case <-timeout.timer.C:
			go fn()
		case <-timeout.t:
			timeout.timer.Stop()
			return
		}
	}()
	return timeout
}

func ClearTimeout(timeout *Timer) {
	if timeout != nil {
		timeout.mu.Lock()
		defer timeout.mu.Unlock()

		if !timeout.isClose {
			close(timeout.t)
			timeout.isClose = true
		}
	}
}

func SetInterval(fn func(), sleep time.Duration) *Timer {
	timeout := &Timer{
		t:       make(chan struct{}),
		timer:   time.NewTimer(sleep),
		sleep:   sleep,
		isClose: false,
	}
	go func() {
		for {
			select {
			case <-timeout.timer.C:
				timeout.timer.Reset(timeout.sleep)
				go fn()
			case <-timeout.t:
				timeout.timer.Stop()
				return
			}
		}
	}()
	return timeout
}

func ClearInterval(timeout *Timer) {
	ClearTimeout(timeout)
}

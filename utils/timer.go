package utils

import (
	"time"
)

type Timer struct {
	t     chan struct{}
	timer *time.Timer
	sleep time.Duration
	fn    func()
}

func (t *Timer) Refresh() *Timer {
	defer t.timer.Reset(t.sleep)

	if !t.timer.Stop() {
		go t.fn()
	}

	return t
}

func SetTimeOut(fn func(), sleep time.Duration) *Timer {
	timeout := &Timer{
		t:     make(chan struct{}),
		timer: time.NewTimer(sleep),
		sleep: sleep,
	}
	timeout.fn = func() {
		select {
		case <-timeout.timer.C:
			go fn()
		case <-timeout.t:
			return
		}
	}
	go timeout.fn()
	return timeout
}

func ClearTimeout(timeout *Timer) {
	if timeout != nil && timeout.timer.Stop() {
		timeout.t <- struct{}{}
	}
}

func SetInterval(fn func(), sleep time.Duration) *Timer {
	timeout := &Timer{
		t:     make(chan struct{}),
		timer: time.NewTimer(sleep),
		sleep: sleep,
	}
	timeout.fn = func() {
		for {
			select {
			case <-timeout.timer.C:
				timeout.timer.Reset(timeout.sleep)
				go fn()
			case <-timeout.t:
				return
			}
		}
	}
	go timeout.fn()
	return timeout
}

func ClearInterval(timeout *Timer) {
	ClearTimeout(timeout)
}

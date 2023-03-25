package utils

import (
	"runtime"
	"time"
)

type Timer struct {
	t     chan struct{}
	stop  chan struct{}
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

func (t *Timer) Unref() {
	runtime.SetFinalizer(t, func(t *Timer) {
		if t.timer.Stop() {
			t.stop <- struct{}{}
		}
	})
}

func SetTimeOut(fn func(), sleep time.Duration) *Timer {
	timeout := &Timer{
		t:     make(chan struct{}),
		stop:  make(chan struct{}),
		timer: time.NewTimer(sleep),
		sleep: sleep,
	}
	timeout.fn = func() {
		select {
		case <-timeout.timer.C:
			go fn()
		case <-timeout.t:
			return
		case <-timeout.stop:
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
		stop:  make(chan struct{}),
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
			case <-timeout.stop:
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

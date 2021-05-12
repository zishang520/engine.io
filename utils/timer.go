package utils

import (
	"time"
)

type Timer chan struct{}

func SetTimeOut(fn func(), sleep time.Duration) *Timer {
	timer := make(Timer)
	go func() {
		select {
		case <-time.After(sleep):
			fn()
		case <-timer:
			return
		}
	}()
	return &timer
}

func SetInterval(fn func(), sleep time.Duration) *Timer {
	timer := make(Timer)
	go func() {
		for {
			select {
			case <-time.After(sleep):
				fn()
			case <-timer:
				return
			}
		}
	}()
	return &timer
}

func ClearInterval(timer *Timer) {
	ClearTimeOut(timer)
}

func ClearTimeOut(timer *Timer) {
	if timer != nil {
		close(*timer)
		timer = nil
	}
}

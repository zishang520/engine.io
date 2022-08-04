package utils

import (
	"github.com/zishang520/engine.io/log"
)

var _log = log.NewLog()

func init() {
	_log.SetFlags(0)
	_log.SetPrefix("engine")
}

func Log() *log.Log {
	return _log
}

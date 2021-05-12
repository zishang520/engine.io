package types

import (
	"github.com/zishang520/engine.io/events"
)

type Socket interface {
	events.EventEmitter
}

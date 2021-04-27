package types

import (
	"github.com/zishang520/engine.io/events"
)

type Server interface {
	events.EventEmmiter
}

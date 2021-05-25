package types

import (
	"github.com/zishang520/engine.io/events"
)

type Server interface {
	events.EventEmitter

	Clients() map[string]Socket
	ClientsCount() uint64
	Close(...Fn)
	HandleRequest()
	handleUpgrade()
	Attach()
	GenerateId()
}

package types

import (
	"github.com/zishang520/engine.io/events"
)

type Socket interface {
	events.EventEmitter

	ID() string
	Server() Server
	Request() *HttpContext
	Upgraded() bool
	ReadyState() bool
	Transport() Transport
	Send(io.Reader, *packet.Option, func(...interface{})) Socket
	Close(bool)
}

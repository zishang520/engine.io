package types

import (
	"github.com/zishang520/engine.io/events"
	"io"
)

type Transport interface {
	events.EventEmitter

	UpgradesTo() *Set
	Discard()
	OnRequest(*HttpContext)
	DoClose(Fn)
	Close(...Fn)
	OnError(string, ...string)
	OnPacket(*packet.Packet)
	OnData(io.Reader)
	OnClose()
}

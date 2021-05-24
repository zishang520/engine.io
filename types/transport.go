package types

import (
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"io"
)

type Transport interface {
	events.EventEmitter

	Discard()
	OnRequest(*HttpContext)
	DoClose(Fn)
	Close(...Fn)
	OnError(string, ...string)
	OnPacket(*packet.Packet)
	OnData(io.Reader)
	OnClose()
}

package transports

import (
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
)

type Transport interface {
	events.EventEmitter

	SetSid(string)
	SetSupportsBinary(bool)
	SetMaxHttpBufferSize(int64)
	SetGttpCompression(*types.HttpCompression)
	SetPerMessageDeflate(*types.PerMessageDeflate)
	SetReadyState(string)
	Parser() parser.Parser
	Sid() string
	Protocol() int
	Name() string
	SupportsFraming() bool
	HandlesUpgrades() bool
	MaxHttpBufferSize() int64
	HttpCompression() *types.HttpCompression
	PerMessageDeflate() *types.PerMessageDeflate
	ReadyState() string
	Writable() bool
	Discard()
	OnRequest(*types.HttpContext)
	DoClose(types.Callable)
	OnError(string, ...string)
	OnPacket(*packet.Packet)
	OnData(types.BufferInterface)
	OnClose()
	Send([]*packet.Packet)
	Close(...types.Callable)
}

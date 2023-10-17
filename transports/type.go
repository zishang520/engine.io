package transports

import (
	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io-go-parser/parser"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/events"
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
	SetWritable(bool)

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

	// Flags the transport as discarded.
	Discard()
	GetDiscarded() bool

	// Called with an incoming HTTP request.
	OnRequest(*types.HttpContext)

	// Closes the transport.
	DoClose(types.Callable)

	// Called with a transport error.
	OnError(string, error)

	// Called with parsed out a packets from the data stream.
	OnPacket(*packet.Packet)

	// Called with the encoded packet data.
	OnData(_types.BufferInterface)

	// Called upon transport close.
	OnClose()

	// Writes a packet payload.
	Send([]*packet.Packet)

	// Closes the transport.
	Close(...types.Callable)
}

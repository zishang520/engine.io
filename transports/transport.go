package transports

import (
	"sync/atomic"

	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io-go-parser/parser"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/v2/errors"
	"github.com/zishang520/engine.io/v2/events"
	"github.com/zishang520/engine.io/v2/log"
	"github.com/zishang520/engine.io/v2/types"
)

var transport_log = log.NewLog("engine:transport")

type transport struct {
	events.EventEmitter

	// Prototype interface, used to implement interface method rewriting
	_proto_ Transport

	maxHttpBufferSize int64
	httpCompression   *types.HttpCompression
	perMessageDeflate *types.PerMessageDeflate

	sid      string
	protocol int // 3

	_readyState atomic.Value //"open";

	_discarded atomic.Bool // false;

	parser parser.Parser // parser.PaserV3;

	req atomic.Pointer[types.HttpContext]

	supportsBinary bool

	_writable atomic.Bool
}

func MakeTransport() Transport {
	t := &transport{
		EventEmitter: events.New(),
	}
	t._readyState.Store("open")

	t.Prototype(t)

	return t
}

func NewTransport(ctx *types.HttpContext) Transport {
	t := MakeTransport()

	t.Construct(ctx)

	return t
}

func (t *transport) Prototype(_t Transport) {
	t._proto_ = _t
}

func (t *transport) Proto() Transport {
	return t._proto_
}

func (t *transport) Sid() string {
	return t.sid
}

func (t *transport) SetSid(sid string) {
	t.sid = sid
}

func (t *transport) Writable() bool {
	return t._writable.Load()
}

func (t *transport) SetWritable(writable bool) {
	t._writable.Store(writable)
}

func (t *transport) Protocol() int {
	return t.protocol
}

func (t *transport) Discarded() bool {
	return t._discarded.Load()
}

func (t *transport) Parser() parser.Parser {
	return t.parser
}

func (t *transport) Req() *types.HttpContext {
	return t.req.Load()
}

func (t *transport) SetReq(req *types.HttpContext) {
	t.req.Store(req)
}

func (t *transport) SupportsBinary() bool {
	return t.supportsBinary
}

func (t *transport) SetSupportsBinary(supportsBinary bool) {
	t.supportsBinary = supportsBinary
}

func (t *transport) ReadyState() string {
	if v, ok := t._readyState.Load().(string); ok {
		return v
	}
	return ""
}

func (t *transport) SetReadyState(state string) {
	transport_log.Debug(`readyState updated from %s to %s (%s)`, t.ReadyState(), state, t._proto_.Name())

	t._readyState.Store(state)
}

func (t *transport) HttpCompression() *types.HttpCompression {
	return t.httpCompression
}

func (t *transport) SetHttpCompression(httpCompression *types.HttpCompression) {
	t.httpCompression = httpCompression

}
func (t *transport) PerMessageDeflate() *types.PerMessageDeflate {
	return t.perMessageDeflate
}

func (t *transport) SetPerMessageDeflate(perMessageDeflate *types.PerMessageDeflate) {
	t.perMessageDeflate = perMessageDeflate
}

func (t *transport) MaxHttpBufferSize() int64 {
	return t.maxHttpBufferSize
}

func (t *transport) SetMaxHttpBufferSize(maxHttpBufferSize int64) {
	t.maxHttpBufferSize = maxHttpBufferSize
}

// Transport Construct.
func (t *transport) Construct(ctx *types.HttpContext) {
	if eio, ok := ctx.Query().Get("EIO"); ok && eio == "4" {
		t.parser = parser.Parserv4()
	} else {
		t.parser = parser.Parserv3()
	}

	t.protocol = t.parser.Protocol()
	t.supportsBinary = !ctx.Query().Has("b64")
}

// Flags the transport as discarded.
func (t *transport) Discard() {
	t._discarded.Store(true)
}

// Called with an incoming HTTP request.
func (t *transport) OnRequest(req *types.HttpContext) {
	transport_log.Debug("setting request")
	t.SetReq(req)
}

// Closes the transport.
func (t *transport) Close(fn ...types.Callable) {
	if "closed" == t.ReadyState() || "closing" == t.ReadyState() {
		return
	}
	t.SetReadyState("closing")
	fn = append(fn, nil)
	t._proto_.DoClose(fn[0])
}

// Called with a transport error.
func (t *transport) OnError(msg string, desc error) {
	if t.ListenerCount("error") > 0 {
		t.Emit("error", errors.NewTransportError(msg, desc).Err())
	} else {
		transport_log.Debug("ignored transport error %s (%s)", msg, desc)
	}
}

// Called with parsed out a packets from the data stream.
func (t *transport) OnPacket(packet *packet.Packet) {
	t.Emit("packet", packet)
}

// Called with the encoded packet data.
func (t *transport) OnData(data _types.BufferInterface) {
	p, _ := t.parser.DecodePacket(data)
	t.OnPacket(p)
}

// Called upon transport close.
func (t *transport) OnClose() {
	t.SetReadyState("closed")
	t.Emit("close")
}

func (t *transport) HandlesUpgrades() bool {
	return false
}

// Advertise framing support.
func (t *transport) SupportsFraming() bool {
	return false
}

// The name of the transport.
func (t *transport) Name() string {
	return ""
}

// Sends an array of packets.
func (t *transport) Send([]*packet.Packet) {
	transport_log.Debug("Not implemented")
}

func (t *transport) DoClose(types.Callable) {
	transport_log.Debug("Not implemented")
}

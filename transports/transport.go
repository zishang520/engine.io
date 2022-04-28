package transports

import (
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"time"
)

type transport struct {
	events.EventEmitter

	maxHttpBufferSize int64
	httpCompression   *types.HttpCompression
	perMessageDeflate *types.PerMessageDeflate

	sid          string
	protocol     int // 3
	closeTimeout time.Duration

	_readyState string        //"open";
	discarded   bool          // false;
	parser      parser.Parser // parser.PaserV3;

	req            *types.HttpContext
	supportsBinary bool

	// abstruct
	handlesUpgrades bool
	supportsFraming bool
	name            string
	writable        bool

	send    func([]*packet.Packet)                                       // abstract
	doClose func(...types.Callable)                                      // abstract
	onData  func(types.BufferInterface)                                  // abstract
	doWrite func(types.BufferInterface, *packet.Options, types.Callable) // abstract
	onClose types.Callable                                               // abstract
}

func NewTransport(ctx *types.HttpContext) *transport {
	t := &transport{}
	return t.New(ctx)
}

func (t *transport) New(ctx *types.HttpContext) *transport {
	t.EventEmitter = events.New()
	t.onData = t.TransportOnData
	t.onClose = t.TransportOnClose

	t.discarded = false
	t.SetReadyState("open")

	if eio, ok := ctx.Query().Get("EIO"); ok && eio == "4" {
		t.parser = parser.Parserv4()
	} else {
		t.parser = parser.Parserv3()
	}
	t.protocol = t.parser.Protocol()

	return t
}

func (t *transport) Parser() parser.Parser {
	return t.parser
}

func (t *transport) SetSid(sid string) {
	t.sid = sid
}

func (t *transport) Sid() string {
	return t.sid
}

func (t *transport) Protocol() int {
	return t.protocol
}

func (t *transport) SetSupportsBinary(supportsBinary bool) {
	t.supportsBinary = supportsBinary
}

func (t *transport) SetMaxHttpBufferSize(maxHttpBufferSize int64) {
	t.maxHttpBufferSize = maxHttpBufferSize
}

func (t *transport) SetGttpCompression(httpCompression *types.HttpCompression) {
	t.httpCompression = httpCompression

}
func (t *transport) SetPerMessageDeflate(perMessageDeflate *types.PerMessageDeflate) {
	t.perMessageDeflate = perMessageDeflate
}

func (t *transport) MaxHttpBufferSize() int64 {
	return t.maxHttpBufferSize
}

func (t *transport) HttpCompression() *types.HttpCompression {
	return t.httpCompression

}
func (t *transport) PerMessageDeflate() *types.PerMessageDeflate {
	return t.perMessageDeflate
}

func (t *transport) Writable() bool {
	return t.writable
}

func (t *transport) DoClose(fn types.Callable) {
	t.doClose(fn)
}

func (t *transport) OnData(data types.BufferInterface) {
	t.onData(data)
}

func (t *transport) OnClose() {
	t.onClose()
}

func (t *transport) DoWrite(data types.BufferInterface, option *packet.Options, fn types.Callable) {
	t.doWrite(data, option, fn)
}

func (t *transport) Send(packets []*packet.Packet) {
	t.send(packets)
}

func (t *transport) ReadyState() string {
	return t._readyState
}

func (t *transport) SetReadyState(state string) {
	utils.Log().Debug(`readyState updated from %s to %s (%s)`, t._readyState, state, t.Name())
	t._readyState = state
}

func (t *transport) CloseTimeout() time.Duration {
	return t.closeTimeout
}

func (t *transport) Name() string {
	return t.name
}

func (t *transport) HandlesUpgrades() bool {
	return t.handlesUpgrades
}

func (t *transport) SupportsFraming() bool {
	return t.supportsFraming
}

func (t *transport) Discard() {
	t.discarded = true
}

func (t *transport) OnRequest(req *types.HttpContext) {
	utils.Log().Debug("setting request")
	t.req = req
}

func (t *transport) Close(fn ...types.Callable) {
	fn = append(fn, types.Noop)
	if "closed" == t.ReadyState() || "closing" == t.ReadyState() {
		return
	}
	t.SetReadyState("closing")
	t.DoClose(fn[0])
}

func (t *transport) OnError(msg string, desc ...string) {
	desc = append(desc, "")
	if t.ListenerCount("error") > 0 {
		err := errors.New(msg)
		err.Type = "TransportError"
		err.Description = desc[0]
		t.Emit("error", err.Err())
	} else {
		utils.Log().Debug("ignored transport error %s (%s)", msg, desc[0])
	}
}

func (t *transport) OnPacket(packet *packet.Packet) {
	t.Emit("packet", packet)
}

func (t *transport) TransportOnData(data types.BufferInterface) {
	p, _ := t.parser.DecodePacket(data)
	t.OnPacket(p)
}

func (t *transport) TransportOnClose() {
	t.SetReadyState("closed")
	t.Emit("close")
}

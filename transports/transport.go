package transports

import (
	"sync"
	"time"

	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
)

var transport_log = log.NewLog("engine:transport")

type transport struct {
	events.EventEmitter

	maxHttpBufferSize int64
	httpCompression   *types.HttpCompression
	perMessageDeflate *types.PerMessageDeflate

	sid          string
	protocol     int // 3
	closeTimeout time.Duration

	_readyState   string //"open";
	mu_readyState sync.RWMutex

	_discarded   bool // false;
	mu_discarded sync.RWMutex

	parser parser.Parser // parser.PaserV3;

	req            *types.HttpContext
	mu_req         sync.RWMutex
	supportsBinary bool

	// abstruct
	handlesUpgrades bool
	supportsFraming bool
	name            string

	_writable   bool
	mu_writable sync.RWMutex

	send    func([]*packet.Packet)                                                 // abstract
	doClose func(...types.Callable)                                                // abstract
	onData  func(types.BufferInterface)                                            // abstract
	doWrite func(types.BufferInterface, *packet.Options, func(*types.HttpContext)) // abstract
	onClose types.Callable                                                         // abstract

	musend sync.Mutex
}

func NewTransport(ctx *types.HttpContext) *transport {
	t := &transport{}
	return t.New(ctx)
}

// Transport New.
func (t *transport) New(ctx *types.HttpContext) *transport {
	t.EventEmitter = events.New()
	t.onData = t.TransportOnData
	t.onClose = t.TransportOnClose

	t.mu_discarded.Lock()
	t._discarded = false
	t.mu_discarded.Unlock()

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
	t.mu_writable.RLock()
	defer t.mu_writable.RUnlock()

	return t._writable
}

func (t *transport) SetWritable(writable bool) {
	t.mu_writable.Lock()
	defer t.mu_writable.Unlock()

	t._writable = writable
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

func (t *transport) DoWrite(data types.BufferInterface, option *packet.Options, fn func(ctx *types.HttpContext)) {
	t.doWrite(data, option, fn)
}

func (t *transport) Send(packets []*packet.Packet) {
	t.musend.Lock()
	defer t.musend.Unlock()

	t.send(packets)
}

func (t *transport) ReadyState() string {
	t.mu_readyState.RLock()
	defer t.mu_readyState.RUnlock()

	return t._readyState
}

func (t *transport) SetReadyState(state string) {
	transport_log.Debug(`readyState updated from %s to %s (%s)`, t._readyState, state, t.Name())
	t.mu_readyState.Lock()
	defer t.mu_readyState.Unlock()

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

// Flags the transport as discarded.
func (t *transport) Discard() {
	t.mu_discarded.Lock()
	defer t.mu_discarded.Unlock()

	t._discarded = true
}

func (t *transport) GetDiscarded() bool {
	t.mu_discarded.RLock()
	defer t.mu_discarded.RUnlock()

	return t._discarded
}

// Called with an incoming HTTP request.
func (t *transport) OnRequest(req *types.HttpContext) {
	transport_log.Debug("setting request")
	t.req = req
}

// Closes the transport.
func (t *transport) Close(fn ...types.Callable) {
	fn = append(fn, types.Noop)
	if "closed" == t.ReadyState() || "closing" == t.ReadyState() {
		return
	}
	t.SetReadyState("closing")
	t.DoClose(fn[0])
}

// Called with a transport error.
func (t *transport) OnError(msg string, desc ...string) {
	desc = append(desc, "")
	if t.ListenerCount("error") > 0 {
		err := errors.New(msg)
		err.Type = "TransportError"
		err.Description = desc[0]
		t.Emit("error", err.Err())
	} else {
		transport_log.Debug("ignored transport error %s (%s)", msg, desc[0])
	}
}

// Called with parsed out a packets from the data stream.
func (t *transport) OnPacket(packet *packet.Packet) {
	t.Emit("packet", packet)
}

// Called with the encoded packet data.
func (t *transport) TransportOnData(data types.BufferInterface) {
	p, _ := t.parser.DecodePacket(data)
	t.OnPacket(p)
}

// Called upon transport close.
func (t *transport) TransportOnClose() {
	t.SetReadyState("closed")
	t.Emit("close")
}

package engineio

import (
	"bytes"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"io"
)

type transport struct {
	events.EventEmitter

	ReadyState string      //"open";
	Discarded  bool        // false;
	Protocol   int         // 3
	Parser     paser.Paser // paser.PaserV3;

	ctx     *types.HttpContext
	DoClose types.Fn
}

func NewTransport(ctx *types.HttpContext) *transport {
	t := &transport{
		EventEmitter: events.New(),
		ReadyState:   "open",
		Discarded:    false,
		ctx:          ctx,
		DoClose:      types.Noop,
	}

	if bytes.Equal(ctx.QueryArgs().Peek("EIO"), []byte("4")) {
		t.Protocol = 4
		t.Parser = paser.PaserV4
	} else {
		t.Protocol = 3
		t.Parser = paser.PaserV3
	}

	return t
}

func (t *transport) Discard() {
	t.Discarded = true
}

func (t *transport) OnRequest(ctx *types.HttpContext) {
	utils.Log.Debug("setting request")
	t.ctx = ctx
}

func (t *transport) _DoClose(func(types.Fn)) {
	t.DoClose = fn
}

func (t *transport) Close(fn ...types.Fn) {
	fn = append(fn, types.Noop)
	if "closed" == t.ReadyState || "closing" == t.ReadyState {
		return
	}
	t.ReadyState = "closing"
	t.DoClose(fn[0])
}

func (t *transport) OnError(msg string, desc ...string) {
	desc = append(desc, "")
	if len(t.Listeners("error")) > 0 {
		err := errors.New(msg)
		err.Type = "TransportError"
		err.Description = desc[0]
		t.Emit("error", err)
	} else {
		utils.Log.Debug("ignored transport error %s (%s)", msg, desc)
	}
}

func (t *transport) OnPacket(packet *packet.Packet) {
	t.Emit("packet", packet)
}

func (t *transport) OnData(data io.Reader) {
	p, _ := t.parser.DecodePacket(data)
	t.OnPacket(p)
}

func (t *transport) OnClose() {
	t.ReadyState = "closed"
	t.Emit("close")
}

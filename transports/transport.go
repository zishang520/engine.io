package engineio

import (
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"io"
	"net/http"
)

type Transport interface {
	events.EventEmitter
}

type transport struct {
	events.EventEmmiter

	ReadyState string      //"open";
	Discarded  bool        // false;
	Protocol   int         // req._query.EIO === "4" ? 4 : 3; // 3rd revision by default
	Parser     paser.Paser // this.protocol === 4 ? parser_v4 : parser_v3;

	req      *http.Request // this.protocol === 4 ? parser_v4 : parser_v3;
	_doClose types.Fn
}

func NewTransport(req *http.Request) Transport {
	t := &transport{}
	t.ReadyState = "open"
	t.Discarded = false
	t.Protocol = 4           // req._query.EIO === "4" ? 4 : 3; // 3rd revision by default
	t.Parser = paser.PaserV4 // this.protocol === 4 ? parser_v4 : parser_v3;
	t._doClose = types.Noop
	return t
}

func (t *transport) Discard() {
	t.Discarded = true
}

func (t *transport) OnRequest(req *http.Request) {
	utils.Log.Debug("setting request")
	t.req = req
}

func (t *transport) DoClose(fn types.Fn) {
	t._doClose = fn
}

func (t *transport) Close(fn ...types.Fn) {
	fn = append(fn, types.Noop)
	if "closed" == t.ReadyState || "closing" == t.ReadyState {
		return
	}
	t.ReadyState = "closing"
	t._doClose(fn[0])
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

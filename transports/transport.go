package engineio

import (
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
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

	req     *http.Request // this.protocol === 4 ? parser_v4 : parser_v3;
	doClose func()
}

func NewTransport(req *http.Request) Transport {
	t := &transport{}
	t.ReadyState = "open"
	t.Discarded = false
	t.Protocol = 4           // req._query.EIO === "4" ? 4 : 3; // 3rd revision by default
	t.Parser = paser.PaserV4 // this.protocol === 4 ? parser_v4 : parser_v3;
	return t
}

func (t *transport) Discard() {
	t.Discarded = true
}

func (t *transport) OnRequest(req *http.Request) {
	utils.Log.Debug("setting request")
	t.req = req
}

func (t *transport) Close(fn) {
	if "closed" == t.ReadyState || "closing" == t.ReadyState {
		return
	}
	t.ReadyState = "closing"
	t.doClose()
}

func (t *transport) OnError(msg, desc) {
	if len(t.Listeners("error")) > 0 {
		// const err = new Error(msg);
		// err.type = "TransportError";
		// err.description = desc;
		t.Emit("error", "err")
	} else {
		utils.Log.Debug("ignored transport error %s (%s)", msg, desc)
	}
}

func (t *transport) OnPacket(packet *packet.Packet) {
	t.Emit("packet", packet)
}

func (t *transport) OnData(data io.Reader) {
	t.OnPacket(t.parser.DecodePacket(data))
}

func (t *transport) OnClose() {
	t.ReadyState = "closed"
	t.Emit("close")
}

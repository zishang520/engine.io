package engineio

import (
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
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

	req     interface{} // this.protocol === 4 ? parser_v4 : parser_v3;
	doClose func()
}

func (t *transport) Discard() {
	t.Discarded = true
}

// func (t *transport) OnRequest(req) {
// 	debug("setting request")
// 	this.req = req
// }

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
		// debug("ignored transport error %s (%s)", msg, desc)
	}
}

func (t *transport) OnPacket(packet *packet.Packet) {
	t.Emit("packet", packet)
}

func (t *transport) OnData(data) {
	t.OnPacket(t.parser.DecodePacket(data))
}

/**
 * Flags the transport as discarded.
 *
 * @api public
 */
func (this *Transport) Discard() {
	this.discarded = true
}

/**
 * Called with an incoming HTTP request.
 *
 * @param {http.IncomingMessage} request
 * @api public
 */

func (this *Transport) OnRequest(req) {
	// debug(`setting request`);
	this.req = req
}

/**
 * Closes the transport.
 *
 * @api public
 */

func (this *Transport) Close(fn) {
	if `closed` == this.readyState || `closing` == this.readyState {
		return
	}

	this.readyState = `closing`
	this.doClose(fn)
}

/**
 * Called with a transport error.
 *
 * @param {String} message error
 * @param {Object} error description
 * @api public
 */

func (this *Transport) OnError(msg string) {
	if len(this.EventEmitter.Listeners(`error`)) > 0 {
		// err.type = `TransportError`;
		// err.description = desc;
		this.EventEmitter.Emit(`error`, errors.New(msg))
	} else {
		// debug(`ignored transport error %s (%s)`, msg, desc)
	}
}

/**
 * Called with parsed out a packets from the data stream.
 *
 * @param {Object} packet
 * @api public
 */

func (this *Transport) OnPacket(packet *types.Packet) {
	this.EventEmitter.Emit(`packet`, packet)
}

/**
 * Called with the encoded packet data.
 *
 * @param {String} data
 * @api public
 */

func (this *Transport) OnData(data io.Reader) {
	this.onPacket(parser.DecodePacket(data))
}

/**
 * Called upon transport close.
 *
 * @api public
 */

func (this *Transport) OnClose() {
	this.readyState = `closed`
	this.EventEmitter.Emit(`close`)
}

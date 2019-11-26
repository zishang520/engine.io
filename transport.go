package engineio

import (
	events "github.com/kataras/go-events"
)

type Transport struct {
	readyState string
	discarded  bool

	EventEmitter events.EventEmmiter
}

func NewTransport(req) *Transport {
	this := &Transport{}

	this.readyState = `open`
	this.discarded = false

	this.EventEmitter = events.New()

	return this
}

/**
 * Noop function.
 *
 * @api private
 */

var noop = func() {}

/**
 * Flags the transport as discarded.
 *
 * @api private
 */
func (this *Transport) discard() {
	this.discarded = true
}

/**
 * Called with an incoming HTTP request.
 *
 * @param {http.IncomingMessage} request
 * @api private
 */

func (this *Transport) onRequest(req) {
	// debug(`setting request`);
	this.req = req
}

/**
 * Closes the transport.
 *
 * @api private
 */

func (this *Transport) close(fn) {
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
 * @api private
 */

func (this *Transport) onError(msg, desc) {
	if this.listeners(`error`).length {
		// var err = new Error(msg);
		// err.type = `TransportError`;
		// err.description = desc;
		this.EventEmitter.Emit(`error`, `err`)
	} else {
		debug(`ignored transport error %s (%s)`, msg, desc)
	}
}

/**
 * Called with parsed out a packets from the data stream.
 *
 * @param {Object} packet
 * @api private
 */

func (this *Transport) onPacket(packet Packet) {
	this.EventEmitter.Emit(`packet`, packet)
}

/**
 * Called with the encoded packet data.
 *
 * @param {String} data
 * @api private
 */

func (this *Transport) onData(data) {
	this.onPacket(parser.decodePacket(data))
}

/**
 * Called upon transport close.
 *
 * @api private
 */

func (this *Transport) onClose(data) {
	this.readyState = `closed`
	this.EventEmitter.Emit(`close`)
}

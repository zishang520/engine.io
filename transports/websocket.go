package transports

import (
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
)

type WebSocket struct {
	*Transport
	Name            string
	HandlesUpgrades bool
	SupportsFraming bool
}

func NewWebSocket(req) {

	this := &WebSocket{NewTransport(req)}

	onHeaders := func(headers) {
		this.EventEmitter.emit(`headers`, headers)
	}
	this.socket = req.websocket
	this.socket.on(`message`, this.onData.bind(this))
	this.socket.once(`close`, this.onClose.bind(this))
	this.socket.on(`error`, this.onError.bind(this))
	this.socket.on(`headers`, onHeaders)
	this.writable = true
	this.perMessageDeflate = null

	/**
	 * Transport name
	 *
	 * @api public
	 */

	this.Name = `websocket`

	/**
	 * Advertise upgrade support.
	 *
	 * @api public
	 */

	this.HandlesUpgrades = true

	/**
	 * Advertise framing support.
	 *
	 * @api public
	 */

	this.SupportsFraming = true

	return this
}

/**
 * Processes the incoming data.
 *
 * @param {String} encoded packet
 * @api public
 */

func (this *WebSocket) OnData(data) {
	debug(`received "%s"`, data)
	this.Transport.OnData(data)
}

/**
 * Writes a packet payload.
 *
 * @param {Array} packets
 * @api public
 */

func (this *WebSocket) Send(packets []*types.Packet) {

	onEnd := func(err error) {
		if err {
			return this.OnError(`write error`)
		}
		this.writable = true
		this.EventEmitter.Emit(`drain`)
	}
	for _, packet := range packets {
		if packet, err := parser.EncodePacket(packet, this.supportsBinary, false); err != nil {
			this.writable = false
			this.socket.Send(data, opts, onEnd)
		}
	}

}

/**
 * Closes the transport.
 *
 * @api public
 */

func (this *WebSocket) DoClose(fn) {
	// debug(`closing`)
	this.socket.Close()
	// fn && fn();
}

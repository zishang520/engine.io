package transports

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

func (this *WebSocket) Send(packets []Packet) {

	onEnd := func(err) {
		if err {
			return this.OnError(`write error`)
		}
		this.writable = true
		this.EventEmitter.Emit(`drain`)
	}
	send := func(data) {
		debug(`writing "%s"`, data)

		// always creates a new object since ws modifies it
		// var opts = {};
		// if (packet.options) {
		//   opts.compress = packet.options.compress;
		// }

		// if (this.perMessageDeflate) {
		//   var len = `string` == typeof data ? Buffer.byteLength(data) : data.length;
		//   if (len < this.perMessageDeflate.threshold) {
		//     opts.compress = false;
		//   }
		// }

		this.writable = false
		this.socket.send(data, opts, onEnd)
	}
	for packet := range packets {
		parser.encodePacket(packet, this.supportsBinary, send)
	}

}

/**
 * Closes the transport.
 *
 * @api public
 */

func (this *WebSocket) doClose(fn) {
	debug(`closing`)
	this.socket.close()
	// fn && fn();
}

package transports

type WebSocket struct {
	Name            string
	HandlesUpgrades bool
	SupportsFraming bool
}

func NewWebSocket(req) {

	this := &WebSocket{}

	Transport.call(this, req)
	onHeaders := func(headers) {
		self.emit(`headers`, headers)
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
 * @api private
 */

func (this *WebSocket) onData(data) {
	debug(`received "%s"`, data)
	Transport.prototype.onData.call(this, data)
}

/**
 * Writes a packet payload.
 *
 * @param {Array} packets
 * @api private
 */

func (this *WebSocket) send(packets []Packet) {

	onEnd := func(err) {
		if err {
			return this.onError(`write error`, err.stack)
		}
		this.writable = true
		this.emit(`drain`)
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
 * @api private
 */

func (this *WebSocket) doClose(fn) {
	debug(`closing`)
	this.socket.close()
	// fn && fn();
}

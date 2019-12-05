package transports

type Polling struct {
	Name string

	closeTimeout      time.Duran
	maxHttpBufferSize int64
	httpCompression   int64
}

func NewPolling(req) *Polling {
	this := &Polling{}
	this.Name = `polling`

	Transport.call(this, req)

	this.closeTimeout = 30 * 1000 * time.Microsecond
	this.maxHttpBufferSize = 0
	this.httpCompression = 0

	return this
}

/**
 * Overrides onRequest.
 *
 * @param {http.IncomingMessage}
 * @api public
 */

func (this *Polling) OnRequest(req) {
	var res = req.res

	if `GET` == req.method {
		this.OnPollRequest(req, res)
	} else if `POST` == req.method {
		this.OnDataRequest(req, res)
	} else {
		res.writeHead(500)
		res.end()
	}
}

/**
 * The client sends a request awaiting for us to send data.
 *
 * @api public
 */

func (this *Polling) OnPollRequest(req, res) {
	if this.req {
		// debug(`request overlap`)
		// assert: this.res, `.req and .res should be (un)set together`
		this.onError(`overlap from client`)
		res.writeHead(500)
		res.end()
		return
	}

	// debug(`setting request`)

	this.req = req
	this.res = res

	onClose := func() {
		this.onError(`poll connection closed prematurely`)
	}

	cleanup := func() {
		req.removeListener(`close`, onClose)
		// this.req = this.res = null;
	}

	req.cleanup = cleanup
	req.on(`close`, onClose)

	this.writable = true
	this.emit(`drain`)

	// if we`re still writable but had a pending close, trigger an empty send
	if this.writable && this.shouldClose {
		debug(`triggering empty send to append close packet`)
		// this.send([map[string]string{ "type": `noop` }]);
	}
}

/**
 * The client sends a request with data.
 *
 * @api public
 */

func (this *Polling) OnDataRequest(req, res) {
	if this.dataReq {
		// assert: this.dataRes, `.dataReq and .dataRes should be (un)set together`
		this.onError(`data request overlap from client`)
		res.writeHead(500)
		res.end()
		return
	}

	var isBinary = `application/octet-stream` == req.headers[`content-type`]

	this.dataReq = req
	this.dataRes = res

	// var chunks = isBinary ? Buffer.concat([]) : ``;

	cleanup := func() {
		req.removeListener(`data`, onData)
		req.removeListener(`end`, onEnd)
		req.removeListener(`close`, onClose)
		// this.dataReq = this.dataRes = chunks = null;
	}

	onClose := func() {
		cleanup()
		this.onError(`data request connection closed prematurely`)
	}

	onData := func(data) {
		var contentLength int
		if isBinary {
			// chunks = Buffer.concat([chunks, data]);
			contentLength = chunks.length
		} else {
			chunks += data
			contentLength = Buffer.byteLength(chunks)
		}

		if contentLength > this.maxHttpBufferSize {
			// chunks = isBinary ? Buffer.concat([]) : ``;
			req.connection.destroy()
		}
	}

	onEnd := func() {
		this.onData(chunks)

		var headers = map[string]string{
			// text/html is required instead of text/plain to avoid an
			// unwanted download dialog on certain user-agents (GH-43)
			`Content-Type`:   `text/html`,
			`Content-Length`: 2,
		}

		res.writeHead(200, this.headers(req, headers))
		res.end(`ok`)
		cleanup()
	}

	req.on(`close`, onClose)
	if !isBinary {
		req.setEncoding(`utf8`)
	}
	req.on(`data`, onData)
	req.on(`end`, onEnd)
}

/**
 * Processes the incoming data payload.
 *
 * @param {String} encoded payload
 * @api public
 */

func (this *Polling) OnData(data) {
	debug(`received "%s"`, data)
	callback := func(packet) {
		if `close` == packet.Type {
			debug(`got xhr close packet`)
			this.onClose()
			return false
		}

		this.onPacket(packet)
	}

	parser.decodePayload(data, callback)
}

/**
 * Overrides onClose.
 *
 * @api public
 */

func (this *Polling) OnClose() {
	if this.writable {
		// close pending poll request
		// this.send([{ type: `noop` }]);
	}
	Transport.prototype.onClose.call(this)
}

/**
 * Writes a packet payload.
 *
 * @param {Object} packet
 * @api public
 */

func (this *Polling) Send(packets) {
	this.writable = false

	if this.shouldClose {
		debug(`appending close packet to payload`)
		// packets.push({ type: `close` });
		this.shouldClose()
		this.shouldClose = null
	}

	if packet, err := parser.EncodePayload(packets, this.supportsBinary, false); err != nil {

	}
	// var compress = packets.some(func (packet) {
	//   return packet.options && packet.options.compress;
	// });
	this.Write(data /*{ compress: compress }*/)
}

/**
 * Writes data as response to poll request.
 *
 * @param {String} data
 * @param {Object} options
 * @api public
 */

func (this *Polling) Write(data, options) {
	// debug(`writing "%s"`, data)
	this.DoWrite(data, options, func() {
		this.req.cleanup()
	})
}

/**
 * Performs the write.
 *
 * @api public
 */

func (this *Polling) DoWrite(data, options, callback) {
	// explicit UTF-8 is required for pages not served under utf
	// var isString = typeof data == `string`;
	// var contentType = isString
	// ? `text/plain; charset=UTF-8`
	// : `application/octet-stream`;

	// var headers = {
	// `Content-Type`: contentType
	// };
	respond := func(data) {
		// headers[`Content-Length`] = `string` == typeof data ? Buffer.byteLength(data) : data.length;
		this.res.writeHead(200, this.headers(this.req, headers))
		this.res.end(data)
		callback()
	}
	if !this.httpCompression || !options.compress {
		respond(data)
		return
	}

	// var len = isString ? Buffer.byteLength(data) : data.length;
	if len < this.httpCompression.threshold {
		respond(data)
		return
	}

	// var encoding = accepts(this.req).encodings([`gzip`, `deflate`]);
	if !encoding {
		respond(data)
		return
	}

	this.compress(data, encoding, func(err, data) {
		if err {
			this.res.writeHead(500)
			this.res.end()
			callback(err)
			return
		}

		headers[`Content-Encoding`] = encoding
		respond(data)
	})

}

/**
 * Compresses data.
 *
 * @api public
 */

func (this *Polling) Compress(data, encoding, callback) {
	debug(`compressing`)

	// var buffers = [];
	// var nread = 0;

	// compressionMethods[encoding](this.httpCompression)
	//   .on(`error`, callback)
	//   .on(`data`, func (chunk) {
	//     buffers.push(chunk);
	//     nread += chunk.length;
	//   })
	//   .on(`end`, func () {
	//     callback(null, Buffer.concat(buffers, nread));
	//   })
	//   .end(data);
}

/**
 * Closes the transport.
 *
 * @api public
 */

func (this *Polling) DoClose(fn) {
	debug(`closing`)
	onClose := func() {
		clearTimeout(closeTimeoutTimer)
		fn()
		this.onClose()
	}
	// var closeTimeoutTimer;

	if this.dataReq {
		debug(`aborting ongoing data request`)
		this.dataReq.destroy()
	}

	if this.writable {
		debug(`transport writable - closing right away`)
		// this.send([{ type: `close` }]);
		onClose()
	} else if this.discarded {
		debug(`transport discarded - closing right away`)
		onClose()
	} else {
		debug(`transport not writable - buffering orderly close`)
		this.shouldClose = onClose
		closeTimeoutTimer = setTimeout(onClose, this.closeTimeout)
	}

}

/**
 * Returns headers for a response.
 *
 * @param {http.IncomingMessage} request
 * @param {Object} extra headers
 * @api public
 */

func (this *Polling) Headers(req, headers) {
	// prevent XSS warnings on IE
	// https://github.com/LearnBoost/socket.io/pull/1333
	var ua = req.headers[`user-agent`]
	// if (ua && (~ua.indexOf(`;MSIE`) || ~ua.indexOf(`Trident/`))) {
	//   headers[`X-XSS-Protection`] = `0`;
	// }
	this.emit(`headers`, headers)
	return headers
}

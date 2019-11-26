package transports

type XHR struct {
}

/**
 * Ajax polling transport.
 *
 * @api public
 */

func NewXHR(req) {
	// Polling.call(this, req);
}

/**
 * Overrides `onRequest` to handle `OPTIONS`..
 *
 * @param {http.IncomingMessage}
 * @api private
 */

func (this *XHR) onRequest(req) {
	if `OPTIONS` == req.method {
		res := req.res
		headers := this.headers(req)
		headers[`Access-Control-Allow-Headers`] = `Content-Type`
		res.writeHead(200, headers)
		res.end()
	} else {
		Polling.prototype.onRequest.call(this, req)
	}
}

/**
 * Returns headers for a response.
 *
 * @param {http.IncomingMessage} request
 * @param {Object} extra headers
 * @api private
 */

func (this *XHR) headers(req, headers) {
	if req.headers.origin {
		headers[`Access-Control-Allow-Credentials`] = `true`
		headers[`Access-Control-Allow-Origin`] = req.headers.origin
	} else {
		headers[`Access-Control-Allow-Origin`] = `*`
	}

	return Polling.prototype.headers.call(this, req, headers)
}

package transports

type XHR struct {
	*Polling
}

/**
 * Ajax polling transport.
 *
 * @api public
 */

func NewXHR(req) *XHR {
	return &XHR{NewPolling(req)}
	// Polling.call(this, req);
}

/**
 * Overrides `onRequest` to handle `OPTIONS`..
 *
 * @param {http.IncomingMessage}
 * @api public
 */

func (this *XHR) OnRequest(req) {
	if `OPTIONS` == req.method {
		res := req.res
		headers := this.headers(req)
		headers[`Access-Control-Allow-Headers`] = `Content-Type`
		res.writeHead(200, headers)
		res.end()
	} else {
		this.Polling.OnRequest(req)
	}
}

/**
 * Returns headers for a response.
 *
 * @param {http.IncomingMessage} request
 * @param {Object} extra headers
 * @api public
 */

func (this *XHR) Headers(req, headers) {
	if req.headers.origin {
		headers[`Access-Control-Allow-Credentials`] = `true`
		headers[`Access-Control-Allow-Origin`] = req.headers.origin
	} else {
		headers[`Access-Control-Allow-Origin`] = `*`
	}

	return this.Polling.Headers(req, headers)
}

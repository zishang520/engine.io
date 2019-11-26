package engineio

import (
	events "github.com/kataras/go-events"
)

type Server struct {
	clients           map[string]string
	clientsCount      int64
	wsEngine          string
	pingTimeout       int64
	pingInterval      int64
	upgradeTimeout    int64
	maxHttpBufferSize int64
	transports        []interface{}
	allowUpgrades     bool
	allowRequest      string
	cookie            string
	cookiePath        string
	cookieHttpOnly    string
	perMessageDeflate string
	httpCompression   string
	initialPacket     string

	EventEmitter events.EventEmmiter
}

/**
 * Server constructor.
 *
 * @param {Object} options
 * @api public
 */

func NewServer(opts interface{}) *Server {
	this := &Server{}

	this.clients = map[string]string{}
	this.clientsCount = 0

	if opts.wsEngine != `` {
		this.wsEngine = opts.wsEngine
	} else {
		this.wsEngine = `ws`
	}

	if opts.pingTimeout > 0 {
		this.pingTimeout = opts.pingTimeout
	} else {
		this.pingTimeout = 5000
	}

	if opts.pingInterval > 0 {
		this.pingInterval = opts.pingInterval
	} else {
		this.pingInterval = 25000
	}

	if opts.upgradeTimeout > 0 {
		this.upgradeTimeout = opts.upgradeTimeout
	} else {
		this.upgradeTimeout = 10000
	}

	if opts.maxHttpBufferSize > 0 {
		this.maxHttpBufferSize = opts.maxHttpBufferSize
	} else {
		this.maxHttpBufferSize = 10E7
	}

	if len(opts.transports) > 0 {
		this.transports = opts.transports
	} else {
		this.transports = []interface{}{}
	}

	this.allowUpgrades = false != opts.allowUpgrades
	this.allowRequest = opts.allowRequest

	this.cookie = opts.cookie
	this.cookiePath = opts.cookiePath
	this.cookieHttpOnly = false != opts.cookieHttpOnly
	this.perMessageDeflate = opts.cookiePath
	this.httpCompression = opts.httpCompression
	this.initialPacket = opts.initialPacket

	// initialize compression options
	for t := range []string{`perMessageDeflate`, `httpCompression`} {
		if this[t].Status && this[t].Threshold == 0 {
			this[t].Threshold = 1024
		}
	}

	this.EventEmitter = events.New()

	this.init()
}

/**
 * Protocol errors mappings.
 */

var (
	UNKNOWN_TRANSPORT    = 0
	UNKNOWN_SID          = 1
	BAD_HANDSHAKE_METHOD = 2
	BAD_REQUEST          = 3
	FORBIDDEN            = 4
)

var errorMessages = map[int]string{
	0: `Transport unknown`,
	1: `Session ID unknown`,
	2: `Bad handshake method`,
	3: `Bad request`,
	4: `Forbidden`,
}

/**
 * Initialize websocket server
 *
 * @api private
 */

func (this *Server) init() {
	// if (!~this.transports.Index(`websocket`)) return;

	if this.ws {
		this.ws.close()
	}

	var wsModule Ws
	switch this.wsEngine {
	case `uws`:
		wsModule = require(`uws`)
	case `ws`:
		wsModule = require(`ws`)
	default:
		panic(`unknown wsEngine`)
	}
	// this.ws = new wsModule.Server({
	//   noServer: true,
	//   clientTracking: false,
	//   perMessageDeflate: this.perMessageDeflate,
	//   maxPayload: this.maxHttpBufferSize
	// });
}

/**
 * Returns a list of available transports for upgrade given a certain transport.
 *
 * @return {Array}
 * @api public
 */

func (this *Server) Upgrades(transport interface{}) (upgrades []interface{}) {
	if !this.allowUpgrades {
		return upgrades
	}
	return transports[transport].upgradesTo
}

/**
 * Verifies a request.
 *
 * @param {http.IncomingMessage}
 * @return {Boolean} whether the request is valid
 * @api private
 */

func (this *Server) verify(req, upgrade, fn) {
	// transport check
	// var transport = req._query.transport;
	// if (!~this.transports.indexOf(transport)) {
	//   debug(`unknown transport "%s"`, transport);
	//   return fn(Server.errors.UNKNOWN_TRANSPORT, false);
	// }

	// // `Origin` header check
	// var isOriginInvalid = checkInvalidHeaderChar(req.headers.origin);
	// if (isOriginInvalid) {
	//   req.headers.origin = null;
	//   debug(`origin header invalid`);
	//   return fn(Server.errors.BAD_REQUEST, false);
	// }

	// // sid check
	// var sid = req._query.sid;
	// if (sid) {
	//   if (!this.clients.hasOwnProperty(sid)) {
	//     debug(`unknown sid "%s"`, sid);
	//     return fn(Server.errors.UNKNOWN_SID, false);
	//   }
	//   if (!upgrade && this.clients[sid].transport.name !== transport) {
	//     debug(`bad request: unexpected transport without upgrade`);
	//     return fn(Server.errors.BAD_REQUEST, false);
	//   }
	// } else {
	//   // handshake is GET only
	//   if (`GET` !== req.method) return fn(Server.errors.BAD_HANDSHAKE_METHOD, false);
	//   if (!this.allowRequest) return fn(null, true);
	//   return this.allowRequest(req, fn);
	// }

	// fn(null, true);
}

/**
 * Prepares a request by processing the query string.
 *
 * @api private
 */

func (this *Server) prepare(req) {
	// try to leverage pre-existing `req._query` (e.g: from connect)
	// if (!req._query) {
	//   req._query = ~req.url.indexOf(`?`) ? qs.parse(parse(req.url).query) : {};
	// }
}

/**
 * Closes all clients.
 *
 * @api public
 */

func (this *Server) Close(req) *Server {
	// debug(`closing all open clients`);
	for client := range this.clients {
		client.close(true)
	}
	if this.ws {
		// debug(`closing webSocketServer`);
		this.ws.close()
		// don't delete this.ws because it can be used again if the http server starts listening again
	}
	return this
}

/**
 * Handles an Engine.IO HTTP request.
 *
 * @param {http.IncomingMessage} request
 * @param {http.ServerResponse|http.OutgoingMessage} response
 * @api public
 */

func (this *Server) HandleRequest(req, res) {
	// debug(`handling "%s" http request "%s"`, req.method, req.url)
	this.prepare(req)
	req.res = res

	this.verify(req, false, func(err, success) {
		if !success {
			sendErrorMessage(req, res, err)
			return
		}

		if req._query.sid {
			// debug(`setting new request for existing client`)
			this.clients[req._query.sid].transport.onRequest(req)
		} else {
			this.handshake(req._query.transport, req)
		}
	})
}

/**
 * Sends an Engine.IO Error Message
 *
 * @param {http.ServerResponse} response
 * @param {code} error code
 * @api private
 */

func (this *Server) sendErrorMessage(req, res, code) {
	// var headers = { 'Content-Type': 'application/json' };

	// var isForbidden = !Server.errorMessages.hasOwnProperty(code);
	// if (isForbidden) {
	//   res.writeHead(403, headers);
	//   res.end(JSON.stringify({
	//     code: Server.errors.FORBIDDEN,
	//     message: code || Server.errorMessages[Server.errors.FORBIDDEN]
	//   }));
	//   return;
	// }
	// if (req.headers.origin) {
	//   headers['Access-Control-Allow-Credentials'] = 'true';
	//   headers['Access-Control-Allow-Origin'] = req.headers.origin;
	// } else {
	//   headers['Access-Control-Allow-Origin'] = '*';
	// }
	// if (res !== undefined) {
	//   res.writeHead(400, headers);
	//   res.end(JSON.stringify({
	//     code: code,
	//     message: Server.errorMessages[code]
	//   }));
	// }
}

/**
 * generate a socket id.
 * Overwrite this method to generate your custom socket id
 *
 * @param {Object} request object
 * @api public
 */

func (this *Server) GenerateId(req) {
	return base64id.generateId()
}

/**
 * Handshakes a new client.
 *
 * @param {String} transport name
 * @param {Object} request object
 * @api private
 */

func (this *Server) handshake(transportName, req) {
	id := this.generateId(req)

	debug(`handshaking client "%s"`, id)

	// try {
	//  transport := new transports[transportName](req);
	// if (`polling` == transportName) {
	//   transport.maxHttpBufferSize = this.maxHttpBufferSize;
	//   transport.httpCompression = this.httpCompression;
	// } else if (`websocket` == transportName) {
	//   transport.perMessageDeflate = this.perMessageDeflate;
	// }

	// if (req._query && req._query.b64) {
	//   transport.supportsBinary = false;
	// } else {
	//   transport.supportsBinary = true;
	// }
	// } catch (e) {
	// debug(`error handshaking to transport "%s"`, transportName);
	sendErrorMessage(req, req.res, Server.errors.BAD_REQUEST)
	return
	// }
	socket = newSocket(id, this, transport, req)

	if false != this.cookie {
		// transport.EventEmitter.On(`headers`, function (headers) {
		//   headers[`Set-Cookie`] = cookieMod.serialize(this.cookie, id,
		//     {
		//       path: this.cookiePath,
		//       httpOnly: this.cookiePath ? this.cookieHttpOnly : false
		//     });
		// });
	}

	transport.onRequest(req)

	this.clients[id] = socket
	this.clientsCount++

	socket.EventEmitter.Once(`close`, func() {
		delete(this.clients[id])
		this.clientsCount--
	})

	this.EventEmitter.Emit(`connection`, socket)
}

/**
 * Handles an Engine.IO HTTP Upgrade.
 *
 * @api public
 */

func (this *Server) HandleUpgrade(req, socket, upgradeHead) {
	this.prepare(req)

	this.verify(req, true, func(err, success) {
		if !success {
			abortConnection(socket, err)
			return
		}

		head := Buffer.from(upgradeHead) // eslint-disable-line node/no-deprecated-api
		upgradeHead = nil

		// delegate to ws
		this.ws.handleUpgrade(req, socket, head, func(conn) {
			this.onWebSocket(req, conn)
		})
	})
}

/**
 * Called upon a ws.io connection.
 *
 * @param {ws.Socket} websocket
 * @api private
 */

func (this *Server) onWebSocket(req, socket) {
	socket.EventEmitter.On(`error`, onUpgradeError)

	if transports[req._query.transport] != undefined && !transports[req._query.transport].prototype.handlesUpgrades {
		debug(`transport doesnt handle upgraded requests`)
		socket.close()
		return
	}

	// get client id
	var id = req._query.sid

	// keep a reference to the ws.Socket
	req.websocket = socket
	onUpgradeError := func() {
		debug(`websocket error before upgrade`)
		// socket.close() not needed
	}
	if id {
		var client = this.clients[id]
		if !client {
			debug(`upgrade attempt for closed client`)
			socket.close()
		} else if client.upgrading {
			debug(`transport has already been trying to upgrade`)
			socket.close()
		} else if client.upgraded {
			debug(`transport had already been upgraded`)
			socket.close()
		} else {
			debug(`upgrading existing transport`)

			// transport error handling takes over
			socket.removeListener(`error`, onUpgradeError)

			var transport = newtransports[req._query.transport](req)
			if req._query && req._query.b64 {
				transport.supportsBinary = false
			} else {
				transport.supportsBinary = true
			}
			transport.perMessageDeflate = this.perMessageDeflate
			client.maybeUpgrade(transport)
		}
	} else {
		// transport error handling takes over
		socket.removeListener(`error`, onUpgradeError)

		this.handshake(req._query.transport, req)
	}

}

/**
 * Captures upgrade requests for a http.Server.
 *
 * @param {http.Server} server
 * @param {Object} options
 * @api public
 */

func (this *Server) Attach(server, options) {
	// var path = (options.path || `/engine.io`).replace(/\/$/, ``);
	// var destroyUpgradeTimeout = options.destroyUpgradeTimeout || 1000;
	// // normalize path
	// path += `/`;

	// check:=func (req) {
	//   if (`OPTIONS` == req.method && false == options.handlePreflightRequest) {
	//     return false;
	//   }
	//   return path == req.url.substr(0, path.length);
	// }

	// // cache and clean up listeners
	// var listeners = server.listeners(`request`).slice(0);
	// server.removeAllListeners(`request`);
	// server.EventEmitter.On(`close`, this.close.bind(this));
	// server.EventEmitter.On(`listening`, this.init.bind(this));

	// // add request handler
	// server.EventEmitter.On(`request`, function (req, res) {
	//   if (check(req)) {
	//     debug(`intercepting request for path "%s"`, path);
	//     if (`OPTIONS` === req.method && `function` === typeof options.handlePreflightRequest) {
	//       options.handlePreflightRequest.call(server, req, res);
	//     } else {
	//       this.handleRequest(req, res);
	//     }
	//   } else {
	//     for (var i = 0, l = listeners.length; i < l; i++) {
	//       listeners[i].call(server, req, res);
	//     }
	//   }
	// });

	// if (~this.transports.indexOf(`websocket`)) {
	//   server.EventEmitter.On(`upgrade`, function (req, socket, head) {
	//     if (check(req)) {
	//       this.handleUpgrade(req, socket, head);
	//     } else if (false !== options.destroyUpgrade) {
	//       // default node behavior is to disconnect when no handlers
	//       // but by adding a handler, we prevent that
	//       // and if no eio thing handles the upgrade
	//       // then the socket needs to die!
	//       setTimeout(function () {
	//         if (socket.writable && socket.bytesWritten <= 0) {
	//           return socket.end();
	//         }
	//       }, destroyUpgradeTimeout);
	//     }
	//   });
	// }
}

/**
 * Closes the connection
 *
 * @param {net.Socket} socket
 * @param {code} error code
 * @api private
 */

func (this *Server) abortConnection(socket, code) {
	if socket.writable {
		// var message = Server.errorMessages.hasOwnProperty(code) ? Server.errorMessages[code] : String(code || ``);
		// var length = Buffer.byteLength(message);
		// socket.write(
		//   `HTTP/1.1 400 Bad Request\r\n` +
		//   `Connection: close\r\n` +
		//   `Content-type: text/html\r\n` +
		//   `Content-Length: ` + length + `\r\n` +
		//   `\r\n` +
		//   message
		// );
	}
	socket.destroy()
}

/* eslint-disable */

/**
 * From https://github.com/nodejs/node/blob/v8.4.0/lib/_http_common.js#L303-L354
 *
 * True if val contains an invalid field-vchar
 *  field-value    = *( field-content / obs-fold )
 *  field-content  = field-vchar [ 1*( SP / HTAB ) field-vchar ]
 *  field-vchar    = VCHAR / obs-text
 *
 * checkInvalidHeaderChar() is currently designed to be inlinable by v8,
 * so take care when making changes to the implementation so that the source
 * code size does not exceed v8's default max_inlined_source_size setting.
 **/
var validHdrChars = [...]bool{
	false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, // 0 - 15
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, // 16 - 31
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 32 - 47
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 48 - 63
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 64 - 79
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 80 - 95
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 96 - 111
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, // 112 - 127
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 128 ...
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // ... 255
}

func checkInvalidHeaderChar(val string) bool {
	length := len(val)
	if length < 1 {
		return false
	}
	if !validHdrChars[byte(val[0:1])] {
		// debug(`invalid header, index 0, char "%s"`, val.charCodeAt(0))
		return true
	}
	if length < 2 {
		return false
	}
	if !validHdrChars[byte(val[1:1])] {
		// debug(`invalid header, index true, char "%s"`, val.charCodeAt(1))
		return true
	}
	if length < 3 {
		return false
	}
	if !validHdrChars[byte(val[2:1])] {
		// debug(`invalid header, index 2, char "%s"`, val.charCodeAt(2))
		return true
	}
	if length < 4 {
		return false
	}
	if !validHdrChars[byte(val[3:1])] {
		// debug(`invalid header, index 3, char "%s"`, val.charCodeAt(3))
		return true
	}
	for i = 4; i < length; i += 1 {
		if !validHdrChars[byte(val[i:1])] {
			// debug(`invalid header, index "%i", char "%s"`, i, val.charCodeAt(i))
			return true
		}
	}
	return false
}

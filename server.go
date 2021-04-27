package engineio

import (
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"net/http"
)

/**
 * Protocol errors mappings.
 */
const (
	OK                           int = -1
	UNKNOWN_TRANSPORT            int = 0
	UNKNOWN_SID                  int = 1
	BAD_HANDSHAKE_METHOD         int = 2
	BAD_REQUEST                  int = 3
	FORBIDDEN                    int = 4
	UNSUPPORTED_PROTOCOL_VERSION int = 5
)

var errorMessages map[int]string = map[int]string{
	OK:                           `Ok`,
	UNKNOWN_TRANSPORT:            `Transport unknown`,
	UNKNOWN_SID:                  `Session ID unknown`,
	BAD_HANDSHAKE_METHOD:         `Bad handshake method`,
	BAD_REQUEST:                  `Bad request`,
	FORBIDDEN:                    `Forbidden`,
	UNSUPPORTED_PROTOCOL_VERSION: "Unsupported protocol version",
}

type Server interface {
	events.EventEmmiter
}

type server struct {
	events.EventEmmiter

	clients      map[string]Socket
	clientsCount uint64
	Opts         *types.Config

	ws interface{}
}

func NewServer(opts *types.Config) *server {
	s := &Server{}

	s.clients = map[string]Socket{}
	s.clientsCount = 0

	s.Opts = types.InitConfig
	s.Opts.Assign(opts)

	if s.Opts.Cors != nil {
		s.corsMiddleware = utils.MiddlewareWrapper(s.Opts.Cors)
	}

	s.init()

	return s
}

/**
 * Initialize websocket server
 *
 * @api private
 */

func (s *server) init() {
	if _, ok := s.Opts.Transports["websocket"]; !ok {
		return
	}

	if s.ws != nil {
		s.ws.close()
	}

	// this.ws = new this.opts.wsEngine({
	//   noServer: true,
	//   clientTracking: false,
	//   perMessageDeflate: this.opts.perMessageDeflate,
	//   maxPayload: this.opts.maxHttpBufferSize
	// });
}

/**
 * Returns a list of available transports for upgrade given a certain transport.
 *
 * @return {Array}
 * @api public
 */

func (s *server) upgrades(transport string) *types.Set {
	if !s.Opts.AllowUpgrades {
		return &types.Set{}
	}
	return transports.Transports[transport].UpgradesTo()
}

/**
 * Verifies a request.
 *
 * @param {http.IncomingMessage}
 * @return {Boolean} whether the request is valid
 * @api private
 */

func (s *server) verify(ctx *types.HttpContext, upgrade bool) (int, bool) {
	// transport check
	transport := ctx.Request.URL.Query().Get("transport")
	if _, ok := s.Opts.Transports[transport]; !ok {
		utils.Log.Debug(`unknown transport "%s"`, transport)
		return UNKNOWN_TRANSPORT, false
	}

	// 'Origin' header check
	if utils.CheckInvalidHeaderChar(ctx.Request.Header.Get("Origin")) {
		ctx.Request.Header.Del("Origin")
		utils.Log.Debug("origin header invalid")
		return BAD_REQUEST, false
	}

	// sid check
	sid := ctx.Request.URL.Query().Get("sid")
	if sid != "" {
		if _, ok := s.clients[sid]; !ok {
			utils.Log.Debug(`unknown sid "%s"`, sid)
			return UNKNOWN_SID, false
		}
		if !upgrade && s.clients[sid].Transport().Name() != transport {
			utils.Log.Debug("bad request: unexpected transport without upgrade")
			return BAD_REQUEST, false
		}
	} else {
		// handshake is GET only
		if "GET" != strings.ToUpper(ctx.Request.Method) {
			return BAD_HANDSHAKE_METHOD, false
		}
		if s.Opts.AllowRequest == nil {
			return OK, true
		}
		return s.Opts.AllowRequest(ctx)
	}

	return OK, true
}

/**
 * Prepares a request by processing the query string.
 *
 * @api private
 */

func (s *server) prepare(ctx *types.HttpContext) {
	// try to leverage pre-existing `req._query` (e.g: from connect)
	// if (!req._query) {
	//   req._query = ~req.url.indexOf("?") ? qs.parse(parse(req.url).query) : {};
	// }
}

/**
 * Closes all clients.
 *
 * @api public
 */

func (s *server) close() {
	utils.Log.Debug("closing all open clients")
	for _, client := range s.clients {
		client.Close()
	}
	if s.ws != nil {
		utils.Log.Debug("closing webSocketServer")
		s.ws.Close()
		// don't delete s.ws because it can be used again if the http server starts listening again
	}
	return s
}

/**
 * Handles an Engine.IO HTTP request.
 *
 * @param {http.IncomingMessage} request
 * @param {http.ServerResponse|http.OutgoingMessage} response
 * @api public
 */

func (s *server) handleRequest(req *http.Request, res http.ResponseWriter) {
	utils.Log.Debug(`handling "%s" http request "%s"`, req.Method, req.RequestURI)
	this.prepare(req)

	ctx := &types.HttpContext{
		Request:  req,
		Response: res,
	}

	callback := func(err int, success bool) {
		if !success {
			s.sendErrorMessage(ctx, err)
			return
		}

		if sid := ctx.Request.URL.Query().Get("sid"); sid != "" {
			utils.Log.Debug("setting new request for existing client")
			s.clients[sid].Transport.OnRequest(ctx)
		} else {
			s.handshake(ctx.Request.URL.Query().Get("transport"), ctx)
		}
	}

	if s.corsMiddleware != nil {
		s.corsMiddleware(ctx, func() {
			err, scuuess := s.verify(ctx, false)
			callback(err, success)
		})
	} else {
		err, scuuess := s.verify(ctx, false)
		callback(err, success)
	}
}

func (s *server) generateId(_ *types.HttpContext) (string, error) {
	return utils.Base64Id.GenerateId()
}

func (s *server) handshake(transportName string, ctx *types.HttpContext) {
	protocol := 3 // 3rd revision by default
	if ctx.Request.URL.Query().Get("EIO") == "4" {
		protocol := 4
	}

	if protocol == 3 && !s.opts.AllowEIO3 {
		utils.Log.Debug("unsupported protocol version")
		s.sendErrorMessage(ctx, UNSUPPORTED_PROTOCOL_VERSION)
		return
	}

	id, err := this.generateId(req)
	if err != nil {
		utils.Log.Debug("error while generating an id")
		s.sendErrorMessage(ctx, BAD_REQUEST)
		return
	}

	utils.Log.Debug(`handshaking client "%s"`, id)

	transport, ok := transports.Transports[transportName]
	if !ok {
		utils.Log.Debug(`error handshaking to transport "%s"`, transportName)
		s.sendErrorMessage(ctx, BAD_REQUEST)
		return
	}

	if "polling" == transportName {
		transport.SetMaxHttpBufferSize(s.opts.MaxHttpBufferSize)
		transport.SetGttpCompression(s.opts.HttpCompression)
	} else if "websocket" == transportName {
		transport.SetPerMessageDeflate(s.opts.PerMessageDeflate)
	}

	if _, ok := ctx.Request.URL.Query()["b64"]; ok {
		transport.SetSupportsBinary(false)
	} else {
		transport.SetSupportsBinary(true)
	}

	socket := NewSocket(id, s, transport, ctx, protocol)

	if s.opts.Cookie != nil {
		transport.On("headers", func(headers) {
			headers["Set-Cookie"] = s.opts.Cookie.String()
		})
	}

	transport.OnRequest(ctx)

	s.clients[id] = socket
	s.clientsCount++

	socket.Once("close", func() {
		delete(s.clients[id])
		s.clientsCount--
	})

	s.Emit("connection", socket)
}

func (s *server) handleUpgrade(req, socket, upgradeHead) {
	// this.prepare(req);

	// const self = this;
	// this.verify(req, true, function(err, success) {
	//   if (!success) {
	//     abortConnection(socket, err);
	//     return;
	//   }

	//   const head = Buffer.from(upgradeHead); // eslint-disable-line node/no-deprecated-api
	//   upgradeHead = null;

	//   // delegate to ws
	//   self.ws.handleUpgrade(req, socket, head, function(conn) {
	//     self.onWebSocket(req, conn);
	//   });
	// });
}

func (s *server) onWebSocket(req, socket) {
	// socket.on("error", onUpgradeError);

	// if (
	//   transports[req._query.transport] !== undefined &&
	//   !transports[req._query.transport].prototype.handlesUpgrades
	// ) {
	//   debug("transport doesnt handle upgraded requests");
	//   socket.close();
	//   return;
	// }

	// // get client id
	// const id = req._query.sid;

	// // keep a reference to the ws.Socket
	// req.websocket = socket;

	// if (id) {
	//   const client = this.clients[id];
	//   if (!client) {
	//     debug("upgrade attempt for closed client");
	//     socket.close();
	//   } else if (client.upgrading) {
	//     debug("transport has already been trying to upgrade");
	//     socket.close();
	//   } else if (client.upgraded) {
	//     debug("transport had already been upgraded");
	//     socket.close();
	//   } else {
	//     debug("upgrading existing transport");

	//     // transport error handling takes over
	//     socket.removeListener("error", onUpgradeError);

	//     const transport = new transports[req._query.transport](req);
	//     if (req._query && req._query.b64) {
	//       transport.supportsBinary = false;
	//     } else {
	//       transport.supportsBinary = true;
	//     }
	//     transport.perMessageDeflate = this.perMessageDeflate;
	//     client.maybeUpgrade(transport);
	//   }
	// } else {
	//   // transport error handling takes over
	//   socket.removeListener("error", onUpgradeError);

	//   this.handshake(req._query.transport, req);
	// }

	// function onUpgradeError() {
	//   debug("websocket error before upgrade");
	//   // socket.close() not needed
	// }
}

/**
 * Captures upgrade requests for a http.Server.
 *
 * @param {http.Server} server
 * @param {Object} options
 * @api public
 */

// func (s *Kernel) ServeHTTP(response http.ResponseWriter, request *http.Request) {
func (s *server) attach(server, options) {
	//   const self = this;
	//   options = options || {};
	//   let path = (options.path || "/engine.io").replace(/\/$/, "");

	//   const destroyUpgradeTimeout = options.destroyUpgradeTimeout || 1000;

	//   // normalize path
	//   path += "/";

	//   function check(req) {
	//     return path === req.url.substr(0, path.length);
	//   }

	//   // cache and clean up listeners
	//   const listeners = server.listeners("request").slice(0);
	//   server.removeAllListeners("request");
	//   server.on("close", self.close.bind(self));
	//   server.on("listening", self.init.bind(self));

	//   // add request handler
	//   server.on("request", function(req, res) {
	//     if (check(req)) {
	//       debug('intercepting request for path "%s"', path);
	//       self.handleRequest(req, res);
	//     } else {
	//       let i = 0;
	//       const l = listeners.length;
	//       for (; i < l; i++) {
	//         listeners[i].call(server, req, res);
	//       }
	//     }
	//   });

	//   if (~self.opts.transports.indexOf("websocket")) {
	//     server.on("upgrade", function(req, socket, head) {
	//       if (check(req)) {
	//         self.handleUpgrade(req, socket, head);
	//       } else if (false !== options.destroyUpgrade) {
	//         // default node behavior is to disconnect when no handlers
	//         // but by adding a handler, we prevent that
	//         // and if no eio thing handles the upgrade
	//         // then the socket needs to die!
	//         setTimeout(function() {
	//           if (socket.writable && socket.bytesWritten <= 0) {
	//             return socket.end();
	//           }
	//         }, destroyUpgradeTimeout);
	//       }
	//     });
	//   }
	// }
}

/**
 * Closes the connection
 *
 * @param {net.Socket} socket
 * @param {code} error code
 * @api private
 */

func (this server) sendErrorMessage(ctx, code) {
	// const headers = { "Content-Type": "application/json" };

	// const isForbidden = !Server.errorMessages.hasOwnProperty(code);
	// if (isForbidden) {
	//   res.writeHead(403, headers);
	//   res.end(
	//     JSON.stringify({
	//       code: Server.errors.FORBIDDEN,
	//       message: code || Server.errorMessages[Server.errors.FORBIDDEN]
	//     })
	//   );
	//   return;
	// }
	// if (res !== undefined) {
	//   res.writeHead(400, headers);
	//   res.end(
	//     JSON.stringify({
	//       code: code,
	//       message: Server.errorMessages[code]
	//     })
	//   );
	// }
}

func (this server) abortConnection(socket, code) {
	// socket.on("error", () => {
	//   debug("ignoring error from closed connection");
	// });
	// if (socket.writable) {
	//   const message = Server.errorMessages.hasOwnProperty(code)
	//     ? Server.errorMessages[code]
	//     : String(code || "");
	//   const length = Buffer.byteLength(message);
	//   socket.write(
	//     "HTTP/1.1 400 Bad Request\r\n" +
	//       "Connection: close\r\n" +
	//       "Content-type: text/html\r\n" +
	//       "Content-Length: " +
	//       length +
	//       "\r\n" +
	//       "\r\n" +
	//       message
	//   );
	// }
	socket.destroy()
}

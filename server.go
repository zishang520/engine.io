package engineio

import (
	"encoding/json"
	"github.com/fasthttp/websocket"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"math"
	"net/http"
	"strconv"
)

/**
 * Protocol errors mappings.
 */
const (
	OK_REQUEST                   int = -1
	UNKNOWN_TRANSPORT            int = 0
	UNKNOWN_SID                  int = 1
	BAD_HANDSHAKE_METHOD         int = 2
	BAD_REQUEST                  int = 3
	FORBIDDEN                    int = 4
	UNSUPPORTED_PROTOCOL_VERSION int = 5
)

var errorMessages map[int]string = map[int]string{
	OK_REQUEST:                   `Ok`,
	UNKNOWN_TRANSPORT:            `Transport unknown`,
	UNKNOWN_SID:                  `Session ID unknown`,
	BAD_HANDSHAKE_METHOD:         `Bad handshake method`,
	BAD_REQUEST:                  `Bad request`,
	FORBIDDEN:                    `Forbidden`,
	UNSUPPORTED_PROTOCOL_VERSION: "Unsupported protocol version",
}

type server struct {
	events.EventEmitter

	clients      map[string]Socket
	clientsCount uint64
	Opts         *types.Config

	ws *websocket.FastHTTPUpgrader
}

func NewServer(opts *types.Config) *server {
	s := &Server{
		EventEmitter: events.New(),
	}

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
	if !s.Opts.Transports.Has("websocket") {
		return
	}

	// if s.ws != nil {
	// 	s.ws.Close()
	// }

	s.ws = &websocket.FastHTTPUpgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: s.Opts.PerMessageDeflate != nil,
	}
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
	transport := string(ctx.QueryArgs().Peek("transport"))
	if !s.Opts.Transports.Has(transport) {
		utils.Log.Debug(`unknown transport "%s"`, transport)
		return UNKNOWN_TRANSPORT, false
	}

	// 'Origin' header check
	if utils.CheckInvalidHeaderChar(ctx.Request.Header.Peek("Origin")) {
		ctx.Request.Header.Del("Origin")
		utils.Log.Debug("origin header invalid")
		return BAD_REQUEST, false
	}

	// sid check
	sid := string(ctx.QueryArgs().Peek("sid"))
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
		if "GET" != strings.ToUpper(string(ctx.Method())) {
			return BAD_HANDSHAKE_METHOD, false
		}
		if s.Opts.AllowRequest == nil {
			return OK_REQUEST, true
		}
		return s.Opts.AllowRequest(ctx)
	}

	return OK_REQUEST, true
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
		// s.ws.Close()
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

func (s *server) handleRequest(ctx *types.HttpContext) {
	utils.Log.Debug(`handling "%s" http request "%s"`, req.Method, req.RequestURI)

	callback := func(err int, success bool) {
		if !success {
			s.sendErrorMessage(ctx, err)
			return
		}

		if sid := string(ctx.QueryArgs().Peek("sid")); sid != "" {
			utils.Log.Debug("setting new request for existing client")
			s.clients[sid].Transport.OnRequest(ctx)
		} else {
			s.handshake(string(ctx.QueryArgs().Peek("transport")), ctx)
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

func (s *server) generateId(ctx *types.HttpContext) (string, error) {
	return utils.Base64Id.GenerateId(ctx)
}

func (s *server) handshake(transportName string, ctx *types.HttpContext) {
	protocol := 3 // 3rd revision by default
	if string(ctx.QueryArgs().Peek("EIO")) == "4" {
		protocol := 4
	}

	if protocol == 3 && !s.opts.AllowEIO3 {
		utils.Log.Debug("unsupported protocol version")
		s.sendErrorMessage(ctx, UNSUPPORTED_PROTOCOL_VERSION)
		return
	}

	id, err := s.generateId(req)
	if err != nil {
		utils.Log.Debug("error while generating an id")
		s.sendErrorMessage(ctx, BAD_REQUEST)
		return
	}

	utils.Log.Debug(`handshaking client "%s"`, id)

	_transport, ok := transports.Transports[transportName]
	if !ok {
		utils.Log.Debug(`error handshaking to transport "%s"`, transportName)
		s.sendErrorMessage(ctx, BAD_REQUEST)
		return
	}
	transport := _transport.New(ctx)
	if "polling" == transportName {
		transport.SetMaxHttpBufferSize(s.opts.MaxHttpBufferSize)
		transport.SetGttpCompression(s.opts.HttpCompression)
	} else if "websocket" == transportName {
		transport.SetPerMessageDeflate(s.opts.PerMessageDeflate)
	}

	if ctx.QueryArgs().Has("b64") {
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
	atomic.AddUint64(&s.clientsCount, 1)

	socket.Once("close", func() {
		delete(s.clients, id)
		atomic.AddUint64(&s.clientsCount, 1&math.MaxUint32)
	})

	s.Emit("connection", socket)
}

func (s *server) handleUpgrade(ctx *types.HttpContext) {
	code, success := s.verify(ctx, true)
	if !success {
		s.abortConnection(ctx, code)
		return
	}

	// delegate to ws
	conn, err := s.ws.Upgrade(ctx.RequestCtx, func(*websocket.Conn) {
		conn.SetReadLimit(*s.Opts.MaxHttpBufferSize)
		s.onWebSocket(ctx, &types.WebSocketConn{EventEmitter: events.New(), Conn: conn})
	})
	if err != nil {
		utils.Log.Debug("websocket error before upgrade")
	}
}

func (s *server) onWebSocket(ctx *types.HttpContext, socket *types.WebSocketConn) {
	// onUpgradeError := func() {
	// 	utils.Log.Debug("websocket error before upgrade")
	// 	// socket.close() not needed
	// }

	defer func() {
		if recover() != nil {
			utils.Log.Debug("websocket error before upgrade")
		}
	}()

	// socket.On("error", onUpgradeError)

	transportName := string(ctx.QueryArgs().Peek("transport"))

	if transport, ok := transports.Transports[transportName]; ok && !transport.HandlesUpgrades {
		utils.Log.Debug("transport doesnt handle upgraded requests")
		socket.Close()
		return
	}

	// get client id
	id := string(ctx.QueryArgs().Peek("sid"))

	// keep a reference to the ws.Socket
	ctx.Websocket = socket

	if id != "" {
		client, ok := s.clients[id]
		if !ok {
			utils.Log.Debug("upgrade attempt for closed client")
			socket.Close()
		} else {
			if client.Upgrading() {
				utils.Log.Debug("transport has already been trying to upgrade")
				socket.Close()
			} else if client.Upgraded() {
				utils.Log.Debug("transport had already been upgraded")
				socket.Close()
			} else {
				utils.Log.Debug("upgrading existing transport")

				// transport error handling takes over
				// socket.RemoveListener("error", onUpgradeError)

				transport := transports.Transports[transportName].New(ctx)

				if ctx.QueryArgs().Has("b64") {
					transport.SetSupportsBinary(false)
				} else {
					transport.SetSupportsBinary(true)
				}

				transport.SetPerMessageDeflate(s.Opts.PerMessageDeflate)
				client.maybeUpgrade(transport)
			}
		}
	} else {
		// transport error handling takes over
		// socket.RemoveListener("error", onUpgradeError)

		s.handshake(transportName, ctx)
	}
}

/**
 * Captures upgrade requests for a HttpServer.
 *
 * @param {HttpServer} server
 * @param {Object} options
 * @api public
 */

// func (s *Kernel) ServeHTTP(response http.ResponseWriter, request *http.Request) {
func (s *server) Attach(server *HttpServer, options *types.Config) {
	if options == nil {
		options = s.Opts
	}
	path := "/engine.io"
	if options.Path != nil && *options.Path != "" {
		path = strings.TrimRight(*options.Path, "/")
	}

	destroyUpgradeTimeout := 1000 * time.Millisecond
	if options.DestroyUpgradeTimeout != nil {
		destroyUpgradeTimeout = *options.DestroyUpgradeTimeout
	}

	// normalize path
	path += "/"

	server.On("close", func() {
		s.close()
	})
	server.On("listening", func() {
		s.init()
	})
	server.Handler = func(ctx *fasthttp.RequestCtx) {
		if !websocket.FastHTTPIsWebSocketUpgrade(ctx) {
			if strings.HasPrefix(utils.CleanPath(string(ctx.Path())), path) {
				utils.Log.Debug(`intercepting request for path "%s"`, path)
				s.handleRequest(&types.HttpContext{RequestCtx: ctx})
			} else {
				server.DefaultHandler.ServeHTTP(w, r)
			}
		} else {
			if s.Opts.Transports.Has("websocket") {
				if strings.HasPrefix(utils.CleanPath(string(ctx.Path())), path) {
					s.handleUpgrade(&types.HttpContext{Request: r, Response: w, Context: r.Context()})
				} else if options.DestroyUpgrade {
					// default node behavior is to disconnect when no handlers
					// but by adding a handler, we prevent that
					// and if no eio thing handles the upgrade
					// then the socket needs to die!
					utils.SetTimeOut(func() {
						w.Write(nil)
					}, destroyUpgradeTimeout)
				}
			}
		}
	}
}

/**
 * Closes the connection
 *
 * @param {net.Socket} socket
 * @param {code} error code
 * @api private
 */

func (this server) sendErrorMessage(ctx *types.HttpContext, code int) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if _, isForbidden := errorMessages[code]; !isForbidden {
		ctx.SetStatusCode(403)
		for key, value := range headers {
			ctx.Response.Header.Set(key, value)
		}
		message = errorMessages[FORBIDDEN]
		if code != 0 {
			message = strconv.Itoa(code)
		}
		json.NewEncoder(ctx).Encode(&types.ErrorMessage{
			Code:    FORBIDDEN,
			Message: message,
		})
		return
	}
	ctx.SetStatusCode(400)
	for key, value := range headers {
		ctx.Response.Header.Set(key, value)
	}
	json.NewEncoder(ctx).Encode(&types.ErrorMessage{
		Code:    code,
		Message: errorMessages[code],
	})
}

func (this server) abortConnection(ctx *types.HttpContext, code int) {
	defer func() {
		if recover() != nil {
			utils.Log.Debug("ignoring error from closed connection")
		}
	}()
	message, ok := errorMessages[code]
	if !ok {
		message = strconv.Itoa(code)
	}
	ctx.SetStatusCode(400)
	ctx.SetBodyString(message)
}

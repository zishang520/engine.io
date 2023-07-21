package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io/config"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
)

var server_log = log.NewLog("engine")

func (s *server) CreateTransport(transportName string, ctx *types.HttpContext) (transports.Transport, error) {
	if transport, ok := transports.Transports()[transportName]; ok {
		return transport.New(ctx), nil
	}
	return nil, errors.New("unsupported transportName").Err()
}

// Handles an Engine.IO HTTP request.
func (s *server) HandleRequest(ctx *types.HttpContext) {
	server_log.Debug(`handling "%s" http request "%s"`, ctx.Method(), ctx.Request().RequestURI)

	callback := func(errorCode int, errorContext map[string]any) {
		if errorContext != nil {
			s.Emit("connection_error", &types.ErrorMessage{
				CodeMessage: &types.CodeMessage{
					Code:    errorCode,
					Message: errorMessages[errorCode],
				},
				Req:     ctx,
				Context: errorContext,
			})
			abortRequest(ctx, errorCode, errorContext)
			return
		}

		if sid := ctx.Query().Peek("sid"); sid != "" {
			server_log.Debug("setting new request for existing client")
			if socket, ok := s.clients.Load(sid); ok {
				socket.(Socket).Transport().OnRequest(ctx)
			} else {
				abortRequest(ctx, UNKNOWN_SID, map[string]any{"sid": sid})
			}
		} else {
			if errorCode, errorContext, t := s.Handshake(ctx.Query().Peek("transport"), ctx); t == nil {
				abortRequest(ctx, errorCode, errorContext)
			}
		}
	}

	s._applyMiddlewares(ctx, func(err error) {
		if err != nil {
			callback(BAD_REQUEST, map[string]any{"name": "MIDDLEWARE_FAILURE"})
		} else {
			callback(s.Verify(ctx, false))
		}
	})

	<-ctx.Done()
}

// Handles an Engine.IO HTTP Upgrade.
func (s *server) HandleUpgrade(ctx *types.HttpContext) {
	emitError := func(errorCode int, errorContext map[string]any) {
		s.Emit("connection_error", &types.ErrorMessage{
			CodeMessage: &types.CodeMessage{
				Code:    errorCode,
				Message: errorMessages[errorCode],
			},
			Req:     ctx,
			Context: errorContext,
		})
		abortUpgrade(ctx, errorCode, errorContext)
	}
	callback := func(errorCode int, errorContext map[string]any) {
		if errorContext != nil {
			emitError(errorCode, errorContext)
			return
		}

		wsc := &types.WebSocketConn{EventEmitter: events.New()}

		ws := &websocket.Upgrader{
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			EnableCompression: s.opts.PerMessageDeflate() != nil,
			Error: func(_ http.ResponseWriter, _ *http.Request, _ int, reason error) {
				if websocket.IsUnexpectedCloseError(reason) {
					wsc.Emit("close")
				} else {
					wsc.Emit("error", reason)
				}
			},
			CheckOrigin: func(_ *http.Request) bool {
				// Verified in *server.Verify()
				return true
			},
		}

		// delegate to ws
		if conn, err := ws.Upgrade(ctx.Response(), ctx.Request(), ctx.ResponseHeaders.All()); err == nil {
			conn.SetReadLimit(s.opts.MaxHttpBufferSize())
			wsc.Conn = conn
			s.onWebSocket(ctx, wsc)
		} else {
			emitError(BAD_REQUEST, map[string]any{"name": "UPGRADE_FAILURE"})
			server_log.Debug("websocket error before upgrade: %s", err)
		}
	}

	s._applyMiddlewares(ctx, func(err error) {
		if err != nil {
			callback(BAD_REQUEST, map[string]any{"name": "MIDDLEWARE_FAILURE"})
		} else {
			callback(s.Verify(ctx, true))
		}
	})
}

// Called upon a ws.io connection.
func (s *server) onWebSocket(ctx *types.HttpContext, wsc *types.WebSocketConn) {
	onUpgradeError := func(...any) {
		server_log.Debug("websocket error before upgrade")
		// wsc.close() not needed
	}

	wsc.On("error", onUpgradeError)

	transportName := ctx.Query().Peek("transport")
	if transport, ok := transports.Transports()[transportName]; ok && !transport.HandlesUpgrades {
		server_log.Debug("transport doesnt handle upgraded requests")
		wsc.Close()
		return
	}

	// get client id
	id := ctx.Query().Peek("sid")

	// keep a reference to the ws.Socket
	ctx.Websocket = wsc

	if len(id) > 0 {
		client, ok := s.clients.Load(id)

		if !ok {
			server_log.Debug("upgrade attempt for closed client")
			wsc.Close()
		} else if client.(Socket).Upgrading() {
			server_log.Debug("transport has already been trying to upgrade")
			wsc.Close()
		} else if client.(Socket).Upgraded() {
			server_log.Debug("transport had already been upgraded")
			wsc.Close()
		} else {
			server_log.Debug("upgrading existing transport")

			// transport error handling takes over
			wsc.RemoveListener("error", onUpgradeError)

			transport, err := s.CreateTransport(transportName, ctx)
			if err != nil {
				server_log.Debug("upgrading not existing transport")
				wsc.Close()
			} else {
				if ctx.Query().Has("b64") {
					transport.SetSupportsBinary(false)
				} else {
					transport.SetSupportsBinary(true)
				}
				transport.SetPerMessageDeflate(s.opts.PerMessageDeflate())
				client.(Socket).MaybeUpgrade(transport)
			}
		}
	} else {
		if errorCode, errorContext, t := s.Handshake(transportName, ctx); t == nil {
			abortUpgrade(ctx, errorCode, errorContext)
		}
	}
}

type webTransportHandshake struct {
	Sid string `json:"sid"`
}

func (s *server) OnWebTransportSession(ctx *types.HttpContext, wt *webtransport.Server) {
	ctx.Query().Set("EIO", "4")
	emitError := func(errorCode int, errorContext map[string]any) {
		if errorContext != nil {
			s.Emit("connection_error", &types.ErrorMessage{
				CodeMessage: &types.CodeMessage{
					Code:    errorCode,
					Message: errorMessages[errorCode],
				},
				Req:     ctx,
				Context: errorContext,
			})
			abortUpgrade(ctx, errorCode, errorContext)
			return
		}
	}

	wtc := &types.WebTransportConn{EventEmitter: events.New()}

	session, err := wt.Upgrade(ctx.Response(), ctx.Request())
	if err != nil {
		utils.Log().Debug("upgrading failed: %s", err)
		emitError(BAD_REQUEST, map[string]any{"name": "UPGRADE_FAILURE"})
		return
	}

	wtc.Session = session

	timeout := utils.SetTimeOut(func() {
		server_log.Debug("the client failed to establish a bidirectional stream in the given period")
		session.CloseWithError(0, "")
	}, s.opts.UpgradeTimeout())

	// session.ConnectionState()
	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		utils.Log().Debug("session is closed")
		return
	}
	buf := make([]byte, 32)
	// reading the first packet of the stream
	n, err := stream.Read(buf)
	if err != nil {
		utils.Log().Debug("stream is closed")
		return
	}
	utils.ClearTimeout(timeout)
	handshake := buf[:n]

	ctx.WebTransport = wtc

	// handshake is either
	// "0" => new session
	// '0{"sid":"xxxx"}' => upgrade
	if n == 1 && handshake[0] == '0' {
		if errorCode, errorContext, t := s.Handshake(ctx.Request().Proto, ctx); t == nil {
			abortUpgrade(ctx, errorCode, errorContext)
		}
		return
	}
	var wth webTransportHandshake

	if n > 1 && handshake[0] == '0' {
		if err := json.Unmarshal(handshake[1:], &wth); err != nil {
			server_log.Debug("invalid WebTransport handshake")
			session.CloseWithError(0, "")
			return
		}
	} else {
		server_log.Debug("invalid WebTransport handshake")
		session.CloseWithError(0, "")
		return
	}

	if len(wth.Sid) == 0 {
		server_log.Debug("invalid WebTransport handshake")
		session.CloseWithError(0, "")
		return
	}

	client, ok := s.clients.Load(wth.Sid)

	if !ok {
		server_log.Debug("upgrade attempt for closed client")
		session.CloseWithError(0, "")
	} else if client.(Socket).Upgrading() {
		server_log.Debug("transport has already been trying to upgrade")
		session.CloseWithError(0, "")
	} else if client.(Socket).Upgraded() {
		server_log.Debug("transport had already been upgraded")
		session.CloseWithError(0, "")
	} else {
		server_log.Debug("upgrading existing transport")

		transport, err := s.CreateTransport(ctx.Request().Proto, ctx)
		if err != nil {
			server_log.Debug("upgrading not existing transport")
			session.CloseWithError(0, "")
		} else {
			client.(Socket).MaybeUpgrade(transport)
		}
	}
}

// Captures upgrade requests for a types.HttpServer.
func (s *server) Attach(server *types.HttpServer, opts any) {
	options, _ := opts.(config.AttachOptionsInterface)
	path := s._computePath(options)

	server.On("close", func(...any) {
		s.Close()
	})

	server.HandleFunc(path, s.ServeHTTP)
}

// Captures upgrade requests for a http.Handler, Need to handle server shutdown disconnecting client connections.
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		server_log.Debug(`intercepting request for path "%s"`, utils.CleanPath(r.URL.Path))
		s.HandleRequest(types.NewHttpContext(w, r))
	} else if s.opts.Transports().Has("websocket") {
		s.HandleUpgrade(types.NewHttpContext(w, r))
	} else {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

// Close the HTTP long-polling request
func abortRequest(ctx *types.HttpContext, errorCode int, errorContext map[string]any) {
	server_log.Debug("abortRequest %d", errorCode)
	statusCode := http.StatusBadRequest
	if errorCode == FORBIDDEN {
		statusCode = http.StatusForbidden
	}
	message := errorMessages[errorCode]
	if m, ok := errorContext["message"]; ok {
		message = m.(string)
	}
	ctx.ResponseHeaders.Set("Content-Type", "application/json")
	ctx.SetStatusCode(statusCode)
	if b, err := json.Marshal(types.CodeMessage{Code: errorCode, Message: message}); err == nil {
		ctx.Write(b)
		return
	}
	io.WriteString(ctx, `{"code":400,"message":"Bad request"}`)
}

// Close the WebSocket connection
func abortUpgrade(ctx *types.HttpContext, errorCode int, errorContext map[string]any) {
	ctx.On("error", func(...any) {
		server_log.Debug("ignoring error from closed connection")
	})

	server_log.Debug("abortUpgrade %d", errorCode)
	message := errorMessages[errorCode]
	if m, ok := errorContext["message"]; ok {
		message = m.(string)
	}

	if ctx.Websocket != nil {
		defer ctx.Websocket.Close()
		ctx.Websocket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, message))
	} else if ctx.WebTransport != nil {
		ctx.WebTransport.CloseWithError(http.StatusBadRequest, message)
	} else {
		ctx.SetStatusCode(http.StatusBadRequest)
		io.WriteString(ctx, message)
	}
}

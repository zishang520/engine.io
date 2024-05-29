package engine

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io-go-parser/parser"
	_types "github.com/zishang520/engine.io-go-parser/types"
	"github.com/zishang520/engine.io/v2/config"
	"github.com/zishang520/engine.io/v2/errors"
	"github.com/zishang520/engine.io/v2/events"
	"github.com/zishang520/engine.io/v2/transports"
	"github.com/zishang520/engine.io/v2/types"
	"github.com/zishang520/engine.io/v2/utils"
	webtrans "github.com/zishang520/engine.io/v2/webtransport"
)

type server struct {
	BaseServer
	// Captures upgrade requests for a http.Handler, Need to handle server shutdown disconnecting client connections.
	http.Handler

	httpServer *types.HttpServer
}

// new server.
func MakeServer() Server {
	s := &server{BaseServer: MakeBaseServer()}

	s.Prototype(s)

	return s
}

// create server.
func NewServer(opt any) Server {
	s := MakeServer()

	s.Construct(opt)

	return s
}

func (s *server) SetHttpServer(httpServer *types.HttpServer) {
	s.httpServer = httpServer
}

func (s *server) HttpServer() *types.HttpServer {
	return s.httpServer
}

func (s *server) Init() {
}

func (s *server) Cleanup() {
}

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
			s.emitAbortRequest(ctx, errorCode, errorContext)
			return
		}

		if sid := ctx.Query().Peek("sid"); sid != "" {
			server_log.Debug("setting new request for existing client")
			if socket, ok := s.Clients().Load(sid); ok {
				socket.Transport().OnRequest(ctx)
			} else {
				abortRequest(ctx, UNKNOWN_SID, map[string]any{"sid": sid})
			}
		} else {
			if errorCode, t := s.Handshake(ctx.Query().Peek("transport"), ctx); t == nil {
				abortRequest(ctx, errorCode, nil)
			}
		}
	}

	s.ApplyMiddlewares(ctx, func(err error) {
		if err != nil {
			callback(BAD_REQUEST, map[string]any{"name": "MIDDLEWARE_FAILURE"})
		} else {
			callback(s.Verify(ctx, false))
		}
	})

	// Wait for data to be written to the client.
	<-ctx.Done()
}

// Handles an Engine.IO HTTP Upgrade.
func (s *server) HandleUpgrade(ctx *types.HttpContext) {
	callback := func(errorCode int, errorContext map[string]any) {
		if errorContext != nil {
			s.emitAbortRequest(ctx, errorCode, errorContext)
			return
		}

		wsc := &types.WebSocketConn{EventEmitter: events.New()}

		ws := &websocket.Upgrader{
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			EnableCompression: s.Opts().PerMessageDeflate() != nil,
			Error: func(_ http.ResponseWriter, _ *http.Request, _ int, reason error) {
				if websocket.IsUnexpectedCloseError(reason) {
					wsc.Emit("close")
				} else {
					wsc.Emit("error", reason)
				}
			},
			CheckOrigin: func(*http.Request) bool {
				// Verified in *server.Verify()
				return true
			},
		}

		// delegate to ws
		if conn, err := ws.Upgrade(ctx.Response(), ctx.Request(), ctx.ResponseHeaders.All()); err == nil {
			conn.SetReadLimit(s.Opts().MaxHttpBufferSize())
			wsc.Conn = conn
			s.onWebSocket(ctx, wsc)
		} else {
			s.emitAbortRequest(ctx, BAD_REQUEST, map[string]any{"name": "UPGRADE_FAILURE"})
			server_log.Debug("websocket error before upgrade: %s", err.Error())
		}
	}

	s.ApplyMiddlewares(ctx, func(err error) {
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

	if len(id) == 0 {
		if errorCode, t := s.Handshake(transportName, ctx); t == nil {
			abortUpgrade(ctx, errorCode, nil)
		}
		return
	}

	client, ok := s.Clients().Load(id)

	if !ok {
		server_log.Debug("upgrade attempt for closed client")
		wsc.Close()
	} else if client.Upgrading() {
		server_log.Debug("transport has already been trying to upgrade")
		wsc.Close()
	} else if client.Upgraded() {
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
			transport.SetPerMessageDeflate(s.Opts().PerMessageDeflate())
			client.MaybeUpgrade(transport)
		}
	}
}

type webTransportHandshake struct {
	Sid string `json:"sid" mapstructure:"sid" msgpack:"sid"`
}

func (s *server) OnWebTransportSession(ctx *types.HttpContext, wt *webtransport.Server) {
	if allowRequest := s.Opts().AllowRequest(); allowRequest != nil {
		if err := allowRequest(ctx); err != nil {
			s.emitAbortRequest(ctx, FORBIDDEN, map[string]any{"message": err.Error()})
			return
		}
	}

	session, err := wt.Upgrade(ctx.Response(), ctx.Request())
	if err != nil {
		server_log.Debug("upgrading failed: %s", err.Error())
		s.emitAbortRequest(ctx, BAD_REQUEST, map[string]any{"name": "UPGRADE_FAILURE"})
		return
	}

	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		server_log.Debug("session is closed")
		abortUpgrade(ctx, BAD_REQUEST, nil)
		return
	}

	timeout := utils.SetTimeout(func() {
		server_log.Debug("the client failed to establish a bidirectional stream in the given period")
		session.CloseWithError(0, "")
	}, s.Opts().UpgradeTimeout())

	wtc := webtrans.NewConn(session, stream, true, 0, 0, nil, nil, nil)
	wtc.SetReadLimit(s.Opts().MaxHttpBufferSize())

	ctx.WebTransport = wtc

	mt, message, err := wtc.NextReader()
	if err != nil {
		server_log.Debug("stream is closed: %s", err.Error())
		abortUpgrade(ctx, BAD_REQUEST, nil)
		return
	}

	var data _types.BufferInterface

	switch mt {
	case webtrans.BinaryMessage:
		data = _types.NewBytesBuffer(nil)
		if _, err := data.ReadFrom(message); err != nil {
			server_log.Debug("WebTransport handshake data read failed: %s", err.Error())
			abortUpgrade(ctx, BAD_REQUEST, nil)
			return
		}
	case webtrans.TextMessage:
		data = _types.NewStringBuffer(nil)
		if _, err := data.ReadFrom(message); err != nil {
			server_log.Debug("WebTransport handshake data read failed: %s", err.Error())
			abortUpgrade(ctx, BAD_REQUEST, nil)
			return
		}
	}
	if c, ok := message.(io.Closer); ok {
		c.Close()
	}

	utils.ClearTimeout(timeout)

	value, _ := parser.Parserv4().DecodePacket(data)

	if v, ok := value.Data.(io.Closer); ok {
		defer v.Close()
	}

	if value.Type != packet.OPEN {
		server_log.Debug("invalid WebTransport handshake")
		abortUpgrade(ctx, BAD_REQUEST, nil)
		return
	}

	if data, ok := value.Data.(_types.BufferInterface); ok && data.Len() == 0 {
		ctx.Query().Set("EIO", "4")
		if errorCode, t := s.Handshake(ctx.Request().Proto, ctx); t == nil {
			abortUpgrade(ctx, errorCode, nil)
		}
		return
	}

	var wth *webTransportHandshake
	if json.NewDecoder(value.Data).Decode(&wth) != nil {
		server_log.Debug("invalid WebTransport handshake")
		abortUpgrade(ctx, BAD_REQUEST, nil)
		return
	}

	if len(wth.Sid) == 0 {
		server_log.Debug("invalid WebTransport handshake")
		abortUpgrade(ctx, BAD_REQUEST, nil)
		return
	}

	client, ok := s.Clients().Load(wth.Sid)

	if !ok {
		server_log.Debug("upgrade attempt for closed client")
		session.CloseWithError(0, "")
	} else if client.Upgrading() {
		server_log.Debug("transport has already been trying to upgrade")
		session.CloseWithError(0, "")
	} else if client.Upgraded() {
		server_log.Debug("transport had already been upgraded")
		session.CloseWithError(0, "")
	} else {
		server_log.Debug("upgrading existing transport")

		transport, err := s.CreateTransport(ctx.Request().Proto, ctx)
		if err != nil {
			server_log.Debug("upgrading not existing transport")
			session.CloseWithError(0, "")
		} else {
			transport.SetPerMessageDeflate(s.Opts().PerMessageDeflate())
			client.MaybeUpgrade(transport)
		}
	}
}

// Captures upgrade requests for a types.HttpServer.
func (s *server) Attach(server *types.HttpServer, opts any) {
	options, _ := opts.(config.AttachOptionsInterface)
	path := s.ComputePath(options)

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
	} else if s.Opts().Transports().Has("websocket") {
		s.HandleUpgrade(types.NewHttpContext(w, r))
	} else {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

// Close the HTTP long-polling request
func abortRequest(ctx *types.HttpContext, errorCode int, errorContext map[string]any) {
	server_log.Debug("abortRequest %d, %+v", errorCode, errorContext)
	statusCode := http.StatusBadRequest
	if errorCode == FORBIDDEN {
		statusCode = http.StatusForbidden
	}
	message := errorMessages[errorCode]
	if errorContext != nil {
		if m, ok := errorContext["message"]; ok {
			message = m.(string)
		}
	}
	ctx.ResponseHeaders.Set("Content-Type", "application/json")
	ctx.SetStatusCode(statusCode)
	if b, err := json.Marshal(types.CodeMessage{Code: errorCode, Message: message}); err == nil {
		ctx.Write(b)
		return
	}
	io.WriteString(ctx, `{"code":400,"message":"Bad request"}`)
}

func (s *server) emitAbortRequest(ctx *types.HttpContext, errorCode int, errorContext map[string]any) {
	s.Emit("connection_error", &types.ErrorMessage{
		CodeMessage: &types.CodeMessage{
			Code:    errorCode,
			Message: errorMessages[errorCode],
		},
		Req:     ctx,
		Context: errorContext,
	})
	abortRequest(ctx, errorCode, errorContext)
}

// Close the WebSocket connection
func abortUpgrade(ctx *types.HttpContext, errorCode int, errorContext map[string]any) {
	ctx.On("error", func(...any) {
		server_log.Debug("ignoring error from closed connection")
	})

	message := errorMessages[errorCode]
	if errorContext != nil {
		if m, ok := errorContext["message"]; ok {
			message = m.(string)
		}
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

func (s *server) emitAbortUpgrade(ctx *types.HttpContext, errorCode int, errorContext map[string]any) {
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

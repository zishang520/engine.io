package engine

import (
	"github.com/zishang520/engine.io/config"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"net/http"
	"sync"
	"sync/atomic"
)

// Protocol errors mappings.
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
	OK_REQUEST:                   `OK`,
	UNKNOWN_TRANSPORT:            `Transport unknown`,
	UNKNOWN_SID:                  `Session ID unknown`,
	BAD_HANDSHAKE_METHOD:         `Bad handshake method`,
	BAD_REQUEST:                  `Bad request`,
	FORBIDDEN:                    `Forbidden`,
	UNSUPPORTED_PROTOCOL_VERSION: "Unsupported protocol version",
}

type server struct {
	events.EventEmitter

	clients        *sync.Map
	clientsCount   uint64
	corsMiddleware func(*types.HttpContext, types.Callable)
	opts           *config.ServerOptions

	httpServer *types.HttpServer

	MU sync.RWMutex
}

// Server New.
func NewServer(opt interface{}) *server {
	s := &server{
		EventEmitter: events.New(),
	}

	return s.New(opt)
}

// Server New.
func (s *server) New(opt interface{}) *server {
	opts, _ := opt.(*config.ServerOptions)

	s.clients = &sync.Map{}
	atomic.StoreUint64(&s.clientsCount, 0)

	var err error
	s.opts, err = config.DefaultServerOptions().Assign(opts)
	if err != nil {
		panic(err)
	}

	if opts != nil {
		if cookie := opts.Cookie(); cookie != nil {
			if len(cookie.Name) == 0 {
				cookie.Name = "io"
			}
			if len(cookie.Path) == 0 {
				cookie.Path = "/"
			}
			if len(cookie.Path) > 0 {
				cookie.HttpOnly = true
			}
			if cookie.SameSite == http.SameSiteDefaultMode {
				cookie.SameSite = http.SameSiteLaxMode
			}
			s.opts.SetCookie(cookie)
		}

		if cors := s.opts.Cors(); cors != nil {
			s.corsMiddleware = types.MiddlewareWrapper(cors)
		}
	}

	s.Init()

	return s
}

func (s *server) SetHttpServer(httpServer *types.HttpServer) {
	s.httpServer = httpServer
}

func (s *server) HttpServer() *types.HttpServer {
	return s.httpServer
}

func (s *server) Opts() *config.ServerOptions {
	return s.opts
}

func (s *server) Clients() *sync.Map {
	return s.clients
}

func (s *server) ClientsCount() uint64 {
	return atomic.LoadUint64(&s.clientsCount)
}

// Returns a list of available transports for upgrade given a certain transport.
func (s *server) Upgrades(transport string) *types.Set[string] {
	if !s.opts.AllowUpgrades() {
		return types.NewSet[string]()
	}
	return transports.Transports()[transport].UpgradesTo
}

// Verifies a request.
func (s *server) Verify(ctx *types.HttpContext, upgrade bool) (int, map[string]interface{}) {
	// transport check
	transport := ctx.Query().Peek("transport")
	if !s.opts.Transports().Has(transport) {
		utils.Log().Debug(`unknown transport "%s"`, transport)
		return UNKNOWN_TRANSPORT, map[string]interface{}{"transport": transport}
	}

	// 'Origin' header check
	if utils.CheckInvalidHeaderChar(ctx.Headers().Get("Origin")) {
		origin := ctx.Headers().Get("Origin")
		ctx.Request().Header.Del("Origin")
		utils.Log().Debug("origin header invalid")
		return BAD_REQUEST, map[string]interface{}{"name": "INVALID_ORIGIN", "origin": origin}
	}

	// sid check
	sid := ctx.Query().Peek("sid")
	if len(sid) > 0 {
		scoket, ok := s.clients.Load(sid)
		if !ok {
			utils.Log().Debug(`unknown sid "%s"`, sid)
			return UNKNOWN_SID, map[string]interface{}{"sid": sid}
		}
		if previousTransport := scoket.(Socket).Transport().Name(); !upgrade && previousTransport != transport {
			utils.Log().Debug("bad request: unexpected transport without upgrade")
			return BAD_REQUEST, map[string]interface{}{"name": "TRANSPORT_MISMATCH", "transport": transport, "previousTransport": previousTransport}
		}
	} else {
		// handshake is GET only
		if method := ctx.Method(); http.MethodGet != method {
			return BAD_HANDSHAKE_METHOD, map[string]interface{}{"method": method}
		}

		if transport == "websocket" && !upgrade {
			utils.Log().Debug("invalid transport upgrade")
			return BAD_REQUEST, map[string]interface{}{"name": "TRANSPORT_HANDSHAKE_ERROR"}
		}

		if s.opts.AllowRequest() == nil {
			return OK_REQUEST, nil
		}

		return s.opts.AllowRequest()(ctx)
	}

	return OK_REQUEST, nil
}

// Closes all clients.
func (s *server) Close() Server {
	utils.Log().Debug("closing all open clients")
	s.clients.Range(func(_, client interface{}) bool {
		client.(Socket).Close(true)
		return true
	})
	s.Cleanup()
	return s
}

// generate a socket id.
// Overwrite this method to generate your custom socket id
func (s *server) GenerateId(*types.HttpContext) (string, error) {
	return utils.Base64Id().GenerateId()
}

// Handshakes a new client.
func (s *server) Handshake(transportName string, ctx *types.HttpContext) (int, map[string]interface{}, transports.Transport) {
	protocol := 3 // 3rd revision by default
	if ctx.Query().Peek("EIO") == "4" {
		protocol = 4
	}

	if protocol == 3 && !s.opts.AllowEIO3() {
		utils.Log().Debug("unsupported protocol version")
		s.Emit("connection_error", &types.ErrorMessage{
			CodeMessage: &types.CodeMessage{
				Code:    UNSUPPORTED_PROTOCOL_VERSION,
				Message: errorMessages[UNSUPPORTED_PROTOCOL_VERSION],
			},
			Req: ctx,
			Context: map[string]interface{}{
				"protocol": protocol,
			},
		})
		return UNSUPPORTED_PROTOCOL_VERSION, map[string]interface{}{"protocol": protocol}, nil
	}

	id, err := s.GenerateId(ctx)
	if err != nil {
		utils.Log().Debug("error while generating an id")
		s.Emit("connection_error", &types.ErrorMessage{
			CodeMessage: &types.CodeMessage{
				Code:    BAD_REQUEST,
				Message: errorMessages[BAD_REQUEST],
			},
			Req: ctx,
			Context: map[string]interface{}{
				"name":  "ID_GENERATION_ERROR",
				"error": err,
			},
		})
		return BAD_REQUEST, map[string]interface{}{"name": "ID_GENERATION_ERROR", "error": err}, nil
	}

	utils.Log().Debug(`handshaking client "%s"`, id)

	transport, err := s.CreateTransport(transportName, ctx)
	if err != nil {
		utils.Log().Debug(`error handshaking to transport "%s"`, transportName)
		s.Emit("connection_error", &types.ErrorMessage{
			CodeMessage: &types.CodeMessage{
				Code:    BAD_REQUEST,
				Message: errorMessages[BAD_REQUEST],
			},
			Req: ctx,
			Context: map[string]interface{}{
				"name":  "TRANSPORT_HANDSHAKE_ERROR",
				"error": err,
			},
		})
		return BAD_REQUEST, map[string]interface{}{"name": "TRANSPORT_HANDSHAKE_ERROR", "error": err}, nil
	}
	if "polling" == transportName {
		transport.SetMaxHttpBufferSize(s.opts.MaxHttpBufferSize())
		transport.SetGttpCompression(s.opts.HttpCompression())
	} else if "websocket" == transportName {
		transport.SetPerMessageDeflate(s.opts.PerMessageDeflate())
	}

	if ctx.Query().Has("b64") {
		transport.SetSupportsBinary(false)
	} else {
		transport.SetSupportsBinary(true)
	}

	socket := NewSocket(id, s, transport, ctx, protocol)

	transport.On("headers", func(args ...interface{}) {
		headers, req := args[0].(map[string]string), args[1].(*types.HttpContext)
		if !ctx.Query().Has("sid") {
			if cookie := s.opts.Cookie(); cookie != nil {
				headers["Set-Cookie"] = cookie.String()
			}
			s.Emit("initial_headers", headers, req)
		}
		s.Emit("headers", headers, req)
	})

	transport.OnRequest(ctx)

	s.clients.Store(id, socket)
	atomic.AddUint64(&s.clientsCount, 1)

	socket.Once("close", func(...interface{}) {
		s.clients.Delete(id)
		atomic.AddUint64(&s.clientsCount, ^uint64(0))
	})

	s.Emit("connection", socket)

	return OK_REQUEST, nil, transport
}

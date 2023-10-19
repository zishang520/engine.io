package engine

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/zishang520/engine.io/config"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/log"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
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

var (
	server_log = log.NewLog("engine")

	errorMessages map[int]string = map[int]string{
		OK_REQUEST:                   `OK`,
		UNKNOWN_TRANSPORT:            `Transport unknown`,
		UNKNOWN_SID:                  `Session ID unknown`,
		BAD_HANDSHAKE_METHOD:         `Bad handshake method`,
		BAD_REQUEST:                  `Bad request`,
		FORBIDDEN:                    `Forbidden`,
		UNSUPPORTED_PROTOCOL_VERSION: "Unsupported protocol version",
	}
)

type baseServer struct {
	// clientsCount has to be first in the struct to guarantee alignment for atomic
	// operations. http://golang.org/pkg/sync/atomic/#pkg-note-BUG
	clientsCount uint64

	events.EventEmitter

	// Prototype interface, used to implement interface method rewriting
	_proto_ BaseServer

	clients     *types.Map[string, Socket]
	middlewares []Middleware
	opts        config.ServerOptionsInterface
}

func MakeBaseServer() BaseServer {
	baseServer := &baseServer{EventEmitter: events.New()}
	baseServer.Prototype(baseServer)
	return baseServer
}

func (bs *baseServer) Prototype(server BaseServer) {
	bs._proto_ = server
}

func (bs *baseServer) Proto() BaseServer {
	return bs._proto_
}

func (bs *baseServer) Opts() config.ServerOptionsInterface {
	return bs.opts
}

func (bs *baseServer) Clients() *types.Map[string, Socket] {
	return bs.clients
}

func (bs *baseServer) ClientsCount() uint64 {
	return atomic.LoadUint64(&bs.clientsCount)
}

func (bs *baseServer) Middlewares() []Middleware {
	return bs.middlewares
}

// BaseServer build.
func (bs *baseServer) Construct(opt any) {
	opts, _ := opt.(config.ServerOptionsInterface)

	bs.clients = &types.Map[string, Socket]{}
	atomic.StoreUint64(&bs.clientsCount, 0)

	bs.opts = config.DefaultServerOptions().Assign(opts)

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
			bs.opts.SetCookie(cookie)
		}

		if cors := bs.opts.Cors(); cors != nil {
			bs.Use(types.MiddlewareWrapper(cors))
		}
	}

	bs._proto_.Init()
}

// abstract
func (bs *baseServer) Init() {
}

// Compute the pathname of the requests that are handled by the server
func (bs *baseServer) ComputePath(options config.AttachOptionsInterface) string {
	path := "/engine.io"

	if options != nil {
		if options.GetRawPath() != nil {
			path = strings.TrimRight(options.Path(), "/")
		}
		if options.AddTrailingSlash() != false {
			// normalize path
			path += "/"
		}
	} else {
		// normalize path
		path += "/"
	}

	return path
}

// Returns a list of available transports for upgrade given a certain transport.
func (bs *baseServer) Upgrades(transport string) *types.Set[string] {
	if !bs.opts.AllowUpgrades() {
		return types.NewSet[string]()
	}
	return transports.Transports()[transport].UpgradesTo
}

// Verifies a request.
func (bs *baseServer) Verify(ctx *types.HttpContext, upgrade bool) (int, map[string]any) {
	// transport check
	transport := ctx.Query().Peek("transport")
	if !bs.opts.Transports().Has(transport) {
		server_log.Debug(`unknown transport "%s"`, transport)
		return UNKNOWN_TRANSPORT, map[string]any{"transport": transport}
	}

	// 'Origin' header check
	if origin := ctx.Headers().Peek("Origin"); utils.CheckInvalidHeaderChar(origin) {
		ctx.Headers().Remove("Origin")
		server_log.Debug("origin header invalid")
		return BAD_REQUEST, map[string]any{"name": "INVALID_ORIGIN", "origin": origin}
	}

	// sid check
	sid := ctx.Query().Peek("sid")
	if len(sid) > 0 {
		scoket, ok := bs.clients.Load(sid)
		if !ok {
			server_log.Debug(`unknown sid "%s"`, sid)
			return UNKNOWN_SID, map[string]any{"sid": sid}
		}
		if previousTransport := scoket.(Socket).Transport().Name(); !upgrade && previousTransport != transport {
			server_log.Debug("bad request: unexpected transport without upgrade")
			return BAD_REQUEST, map[string]any{"name": "TRANSPORT_MISMATCH", "transport": transport, "previousTransport": previousTransport}
		}
	} else {
		// handshake is GET only
		if method := ctx.Method(); http.MethodGet != method {
			return BAD_HANDSHAKE_METHOD, map[string]any{"method": method}
		}

		if transport == "websocket" && !upgrade {
			server_log.Debug("invalid transport upgrade")
			return BAD_REQUEST, map[string]any{"name": "TRANSPORT_HANDSHAKE_ERROR"}
		}

		if allowRequest := bs.opts.AllowRequest(); allowRequest != nil {
			if err := allowRequest(ctx); err != nil {
				return FORBIDDEN, map[string]any{"message": err.Error()}
			}
		}
	}

	return OK_REQUEST, nil
}

// Adds a new middleware.
func (bs *baseServer) Use(fn Middleware) {
	// It seems that there is no need to lock? ? ?
	bs.middlewares = append(bs.middlewares, fn)
}

/**
 * Apply the middlewares to the request.
 */
func (bs *baseServer) ApplyMiddlewares(ctx *types.HttpContext, callback func(error)) {
	if len(bs.middlewares) == 0 {
		server_log.Debug("no middleware to apply, skipping")
		callback(nil)
		return
	}
	var apply func(int)
	apply = func(i int) {
		server_log.Debug("applying middleware n°%d", i+1)
		bs.middlewares[i](ctx, func(err error) {
			if err != nil {
				callback(err)
				return
			}
			if i+1 < len(bs.middlewares) {
				apply(i + 1)
			} else {
				callback(nil)
			}
		})
	}

	apply(0)
}

// Closes all clients.
func (bs *baseServer) Close() BaseServer {
	server_log.Debug("closing all open clients")
	bs.clients.Range(func(_ string, client Socket) bool {
		client.Close(true)
		return true
	})

	bs._proto_.Cleanup()

	return bs
}

func (bs *baseServer) Cleanup() {
}

// generate a socket id.
// Overwrite this method to generate your custom socket id
func (bs *baseServer) GenerateId(*types.HttpContext) (string, error) {
	return utils.Base64Id().GenerateId()
}

// Handshakes a new client.
func (bs *baseServer) Handshake(transportName string, ctx *types.HttpContext) (int, transports.Transport) {
	protocol := 3 // 3rd revision by default
	if ctx.Query().Peek("EIO") == "4" {
		protocol = 4
	}

	if protocol == 3 && !bs.opts.AllowEIO3() {
		server_log.Debug("unsupported protocol version")
		bs.Emit("connection_error", &types.ErrorMessage{
			CodeMessage: &types.CodeMessage{
				Code:    UNSUPPORTED_PROTOCOL_VERSION,
				Message: errorMessages[UNSUPPORTED_PROTOCOL_VERSION],
			},
			Req: ctx,
			Context: map[string]any{
				"protocol": protocol,
			},
		})
		return UNSUPPORTED_PROTOCOL_VERSION, nil
	}

	id, err := bs.GenerateId(ctx)
	if err != nil {
		server_log.Debug("error while generating an id")
		bs.Emit("connection_error", &types.ErrorMessage{
			CodeMessage: &types.CodeMessage{
				Code:    BAD_REQUEST,
				Message: errorMessages[BAD_REQUEST],
			},
			Req: ctx,
			Context: map[string]any{
				"name":  "ID_GENERATION_ERROR",
				"error": err,
			},
		})
		return BAD_REQUEST, nil
	}

	server_log.Debug(`handshaking client "%s"`, id)

	transport, err := bs._proto_.CreateTransport(transportName, ctx)
	if err != nil {
		server_log.Debug(`handshaking client "%s" (%s)`, id, transportName)
		bs.Emit("connection_error", &types.ErrorMessage{
			CodeMessage: &types.CodeMessage{
				Code:    BAD_REQUEST,
				Message: errorMessages[BAD_REQUEST],
			},
			Req: ctx,
			Context: map[string]any{
				"name":  "TRANSPORT_HANDSHAKE_ERROR",
				"error": err,
			},
		})
		return BAD_REQUEST, nil
	}
	if "polling" == transportName {
		transport.SetMaxHttpBufferSize(bs.opts.MaxHttpBufferSize())
		transport.SetHttpCompression(bs.opts.HttpCompression())
	} else if "websocket" == transportName {
		transport.SetPerMessageDeflate(bs.opts.PerMessageDeflate())
	}

	if ctx.Query().Has("b64") {
		transport.SetSupportsBinary(false)
	} else {
		transport.SetSupportsBinary(true)
	}

	socket := NewSocket(id, bs, transport, ctx, protocol)

	transport.On("headers", func(args ...any) {
		headers, req := args[0].(*utils.ParameterBag), args[1].(*types.HttpContext)
		if !ctx.Query().Has("sid") {
			if cookie := bs.opts.Cookie(); cookie != nil {
				headers.Set("Set-Cookie", cookie.String())
			}
			bs.Emit("initial_headers", headers, req)
		}
		bs.Emit("headers", headers, req)
	})

	transport.OnRequest(ctx)

	bs.clients.Store(id, socket)
	atomic.AddUint64(&bs.clientsCount, 1)

	socket.Once("close", func(...any) {
		bs.clients.Delete(id)
		atomic.AddUint64(&bs.clientsCount, ^uint64(0))
	})

	bs.Emit("connection", socket)

	return OK_REQUEST, transport
}

// abstract
func (*baseServer) CreateTransport(string, *types.HttpContext) (transports.Transport, error) {
	return nil, errors.New("CreateTransport interface is not implemented").Err()
}

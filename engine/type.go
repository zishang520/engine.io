package engine

import (
	"io"
	"net/http"

	"github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io/config"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
)

type (
	// Middleware functions are functions that have access to the *types.HttpContext
	// and the next middleware function in the application's context cycle.
	Middleware func(*types.HttpContext, func(error))

	BaseServer interface {
		// #extends

		events.EventEmitter

		// #prototype

		Prototype(BaseServer)
		Proto() BaseServer

		// #getters

		Opts() config.ServerOptionsInterface
		// @protected
		Clients() *types.Map[string, Socket]
		ClientsCount() uint64
		// @protected
		Middlewares() []Middleware

		// #methods

		// Construct() should be called after calling Prototype()
		Construct(any)
		// @protected
		// @abstract
		Init()
		// @protected
		// Compute the pathname of the requests that are handled by the server
		ComputePath(config.AttachOptionsInterface) string
		// Returns a list of available transports for upgrade given a certain transport.
		Upgrades(string) *types.Set[string]
		// @protected
		// Verifies a request.
		Verify(*types.HttpContext, bool) (int, map[string]any)
		// Adds a new middleware.
		Use(Middleware)
		// @protected
		// Apply the middlewares to the request.
		ApplyMiddlewares(*types.HttpContext, func(error))
		// Closes all clients.
		Close() BaseServer
		// @protected
		// @abstract
		Cleanup()
		// generate a socket id.
		// Overwrite this method to generate your custom socket id
		GenerateId(*types.HttpContext) (string, error)
		// @protected
		// Handshakes a new client.
		Handshake(string, *types.HttpContext) (int, transports.Transport)
		// @protected
		// @abstract
		CreateTransport(string, *types.HttpContext) (transports.Transport, error)
	}

	Server interface {
		// #extends

		BaseServer
		// Captures upgrade requests for a http.Handler, Need to handle server shutdown disconnecting client connections.
		http.Handler

		// #setters

		SetHttpServer(*types.HttpServer)

		// #getters

		HttpServer() *types.HttpServer

		// #methods

		CreateTransport(string, *types.HttpContext) (transports.Transport, error)
		// Handles an Engine.IO HTTP request.
		HandleRequest(*types.HttpContext)
		// Handles an Engine.IO HTTP Upgrade.
		HandleUpgrade(*types.HttpContext)
		OnWebTransportSession(*types.HttpContext, *webtransport.Server)
		// Captures upgrade requests for a *types.HttpServer.
		Attach(*types.HttpServer, any)
	}

	Socket interface {
		// #extends

		events.EventEmitter

		// #setters

		SetReadyState(string)

		// #getters

		Protocol() int
		Request() *types.HttpContext
		RemoteAddress() string
		Transport() transports.Transport
		Id() string
		ReadyState() string
		// @private
		Upgraded() bool
		// @private
		Upgrading() bool

		// #methods

		Construct(string, BaseServer, transports.Transport, *types.HttpContext, int)
		// @private
		// Upgrades socket to the given transport
		MaybeUpgrade(transports.Transport)
		// Sends a message packet.
		Send(io.Reader, *packet.Options, func(transports.Transport)) Socket
		Write(io.Reader, *packet.Options, func(transports.Transport)) Socket
		// Closes the socket and underlying transport.
		Close(bool)
	}
)

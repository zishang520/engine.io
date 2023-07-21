package engine

import (
	"io"
	"net/http"
	"sync"

	"github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io-go-parser/packet"
	"github.com/zishang520/engine.io/config"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
)

type Server interface {
	events.EventEmitter

	// Captures upgrade requests for a http.Handler, Need to handle server shutdown disconnecting client connections.
	http.Handler

	SetHttpServer(*types.HttpServer)

	HttpServer() *types.HttpServer
	Opts() config.ServerOptionsInterface
	Clients() *sync.Map
	ClientsCount() uint64

	// Returns a list of available transports for upgrade given a certain transport.
	Upgrades(string) *types.Set[string]

	// Adds a new middleware.
	Use(Middleware)

	// Closes all clients.
	Close() Server

	// Handles an Engine.IO HTTP request.
	HandleRequest(*types.HttpContext)

	// Handles an Engine.IO HTTP Upgrade.
	HandleUpgrade(*types.HttpContext)

	OnWebTransportSession(*types.HttpContext, *webtransport.Server)

	// Captures upgrade requests for a *types.HttpServer.
	Attach(*types.HttpServer, any)

	// generate a socket id.
	// Overwrite this method to generate your custom socket id
	GenerateId(*types.HttpContext) (string, error)
}

type Socket interface {
	events.EventEmitter

	SetReadyState(string)

	Id() string
	ReadyState() string
	Protocol() int
	Server() Server
	Request() *types.HttpContext
	RemoteAddress() string
	Upgraded() bool
	Upgrading() bool
	Transport() transports.Transport

	// Upgrades socket to the given transport
	MaybeUpgrade(transports.Transport)

	// Sends a message packet.
	Send(io.Reader, *packet.Options, func(transports.Transport)) Socket
	Write(io.Reader, *packet.Options, func(transports.Transport)) Socket

	// Closes the socket and underlying transport.
	Close(bool)
}

// Middleware functions are functions that have access to the *types.HttpContext
// and the next middleware function in the application's context cycle.
type Middleware func(*types.HttpContext, func(error))

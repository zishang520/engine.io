package engine

import (
	"github.com/zishang520/engine.io/config"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/transports"
	"github.com/zishang520/engine.io/types"
	"io"
	"sync"
)

type Server interface {
	events.EventEmitter

	SetHttpServer(*types.HttpServer)
	HttpServer() *types.HttpServer
	Opts() *config.ServerOptions
	Clients() *sync.Map
	ClientsCount() uint64
	Upgrades(string) *types.Set
	Close() Server
	HandleRequest(*types.HttpContext)
	HandleUpgrade(*types.HttpContext)
	Attach(*types.HttpServer, interface{})
	GenerateId(*types.HttpContext) (string, error)
}

type Socket interface {
	events.EventEmitter

	ID() string
	Server() Server
	Request() *types.HttpContext
	Upgraded() bool
	Upgrading() bool
	MaybeUpgrade(transports.Transport)
	ReadyState() string
	SetReadyState(string)
	Transport() transports.Transport
	Send(io.Reader, *packet.Options, func(transports.Transport)) Socket
	Close(bool)
}

package engine

import (
	"github.com/zishang520/engine.io/types"
	"net/http"
)

const Protocol = 4

func New(server interface{}, args ...interface{}) Server {
	switch s := server.(type) {
	case *types.HttpServer:
		return Attach(s, args[0])
	case interface{}:
		return NewServer(s)
	}
	return NewServer(nil)
}

func Listen(addr string, options interface{}, fn types.Callable) Server {
	server := types.CreateServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}))

	// create engine server
	engine := Attach(server, options)
	engine.SetHttpServer(server)

	server.Listen(addr, fn)

	return engine
}

func Attach(server *types.HttpServer, options interface{}) Server {
	engine := NewServer(options)
	engine.Attach(server, options)
	return engine
}

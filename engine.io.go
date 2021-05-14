package engineio

import (
	"github.com/valyala/fasthttp"
	"github.com/zishang520/engine.io/events"
	"strings"
)

const Protocol = 1

type HttpServer struct {
	events.EventEmitter
	*fasthttp.Server
	DefaultHandler fasthttp.RequestHandler
}

func createServer(defaultHandler fasthttp.RequestHandler) *HttpServer {
	return &HttpServer{
		EventEmitter:   events.New(),
		Server:         &fasthttp.Server{Handler: defaultHandler},
		DefaultHandler: defaultHandler,
	}
}

func New(server interface{}, arguments ...interface{}) types.Server {
	switch s := server.(type) {
	case *HttpServer:
		if s1, ok := arguments[0]; ok {
			if c, ck := s1.(*types.Config); ck {
				return Attach(s, c)
			}
		}
		return Attach(s, nil)
	case *types.Config:
		return NewServer(s)
	}
	return NewServer(nil)
}

func Listen(addr string, options *types.Config, fn types.Fn) types.Server {
	server := createServer(func(ctx *fasthttp.RequestCtx) {
		ctx.Error("Not Implemented", 501)
	})

	// create engine server
	engine := Attach(server, options)
	engine.HttpServer(server)

	err := server.ListenAndServe(addr)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	server.Emit("listening")

	return engine
}

func Attach(server *HttpServer, options *types.Config) types.Server {
	engine := NewServer(options)
	engine.attach(server, options)
	return engine
}

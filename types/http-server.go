package types

import (
	"github.com/zishang520/engine.io/events"
	"net/http"
)

type HttpServer struct {
	events.EventEmitter
	*ServeMux
}

func CreateServer(defaultHandler http.Handler) *HttpServer {
	if defaultHandler == nil {
		defaultHandler = http.NotFoundHandler()
	}
	s := &HttpServer{
		EventEmitter: events.New(),
		ServeMux:     NewServeMux(),
	}
	s.NotFound = defaultHandler
	return s
}

func (s *HttpServer) server(addr string) *http.Server {
	server := &http.Server{Addr: addr, Handler: s}
	server.RegisterOnShutdown(func() {
		s.Emit("close")
	})
	return server
}

func (s *HttpServer) Listen(addr string, fn Callable) *HttpServer {

	go func() {
		if err := s.server(addr).ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	if fn != nil {
		fn()
	}
	s.Emit("listening")

	return s
}

func (s *HttpServer) ListenTLS(addr string, certFile string, keyFile string, fn Callable) *HttpServer {

	go func() {
		if err := s.server(addr).ListenAndServeTLS(certFile, keyFile); err != nil {
			panic(err)
		}
	}()

	if fn != nil {
		fn()
	}
	s.Emit("listening")

	return s
}

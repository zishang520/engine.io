package types

import (
	"github.com/zishang520/engine.io/events"
	"net/http"
)

type HttpServer struct {
	events.EventEmitter
	*http.Server
	DefaultHandler http.Handler
}

func CreateServer(defaultHandler http.Handler) *HttpServer {
	if defaultHandler == nil {
		defaultHandler = http.NotFoundHandler()
	}
	return &HttpServer{
		EventEmitter:   events.New(),
		Server:         &http.Server{},
		DefaultHandler: defaultHandler,
	}
}

func (s *HttpServer) Listen(addr string, fn Callable) *HttpServer {
	s.Addr = addr

	go func() {
		if err := s.ListenAndServe(); err != nil {
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
	s.Addr = addr

	go func() {
		if err := s.ListenAndServeTLS(certFile, keyFile); err != nil {
			panic(err)
		}
	}()

	if fn != nil {
		fn()
	}
	s.Emit("listening")

	return s
}

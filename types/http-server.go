package types

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/zishang520/engine.io/events"
)

type HttpServer struct {
	events.EventEmitter
	*ServeMux

	servers []*http.Server
	mu      sync.RWMutex
}

func NewWebServer(defaultHandler http.Handler) *HttpServer {
	s := &HttpServer{
		EventEmitter: events.New(),
		ServeMux:     NewServeMux(defaultHandler),
	}
	return s
}

// Deprecated: this method will be removed in the next major release, please use NewWebServer.ListenTLS instead.
func CreateServer(defaultHandler http.Handler) *HttpServer {
	return NewWebServer(defaultHandler)
}

func (s *HttpServer) httpServer(addr string, handler http.Handler) *http.Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	server := &http.Server{Addr: addr, Handler: handler}

	s.servers = append(s.servers, server)

	return server
}

func (s *HttpServer) Close(fn Callable) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.servers != nil {
		s.Emit("close")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, server := range s.servers {
			if err := server.Shutdown(ctx); err != nil {
				return err
			}
		}
		if fn != nil {
			defer fn()
		}
	}
	return nil
}

func (s *HttpServer) Listen(addr string, fn Callable) *HttpServer {
	go func() {
		if err := s.httpServer(addr, s).ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

func (s *HttpServer) ListenTLS(addr string, certFile string, keyFile string, fn Callable) *HttpServer {
	go func() {
		if err := s.httpServer(addr, s).ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

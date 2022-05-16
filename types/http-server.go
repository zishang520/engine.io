package types

import (
	"context"
	"github.com/zishang520/engine.io/events"
	"net/http"
	"sync"
	"time"
)

type HttpServer struct {
	events.EventEmitter
	*ServeMux

	servers []*http.Server
	mu      sync.RWMutex
}

func CreateServer(defaultHandler http.Handler) *HttpServer {
	s := &HttpServer{
		EventEmitter: events.New(),
		ServeMux:     NewServeMux(defaultHandler),
	}
	return s
}

func (s *HttpServer) server(addr string) *http.Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	server := &http.Server{Addr: addr, Handler: s}
	server.RegisterOnShutdown(func() {
		s.Emit("close")
	})

	if s.servers == nil {
		s.servers = []*http.Server{}
	}

	s.servers = append(s.servers, server)

	return server
}

func (s *HttpServer) Close(fn Callable) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.servers != nil {
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
		if err := s.server(addr).ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

// Listen 与 ListenTLS
func (s *HttpServer) ListenTLS(addr string, certFile string, keyFile string, fn Callable) *HttpServer {

	go func() {
		if err := s.server(addr).ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

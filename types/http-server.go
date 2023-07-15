package types

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/zishang520/engine.io/events"
	"golang.org/x/net/http2"
)

type HttpServer struct {
	events.EventEmitter
	*ServeMux

	servers []any
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

	s.servers = append(s.servers, server)

	return server
}

func (s *HttpServer) h3Server() *http3.Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Start the servers
	server := &http3.Server{Handler: s}

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
			switch s := server.(type) {
			case *http.Server:
				if err := s.Shutdown(ctx); err != nil {
					return err
				}
			case *http3.Server:
				if err := s.Close(); err != nil {
					return err
				}
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

func (s *HttpServer) ListenHTTP2TLS(addr string, certFile string, keyFile string, conf *http2.Server, fn Callable) *HttpServer {

	go func() {
		server := s.server(addr)
		if err := http2.ConfigureServer(server, conf); err != nil {
			panic(err)
		}
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

func (s *HttpServer) ListenHTTP3TLS(addr string, certFile string, keyFile string, quicConfig *quic.Config, fn Callable) *HttpServer {

	go func() {

		// Load certs
		var err error
		certs := make([]tls.Certificate, 1)
		certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			panic(err)
		}
		// We currently only use the cert-related stuff from tls.Config,
		// so we don't need to make a full copy.
		config := &tls.Config{
			Certificates: certs,
		}

		if addr == "" {
			addr = ":https"
		}

		// Open the listeners
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			panic(err)
		}
		udpConn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			panic(err)
		}
		defer udpConn.Close()

		server := s.h3Server()
		server.TLSConfig = config
		server.QuicConfig = quicConfig

		hErr := make(chan error)
		qErr := make(chan error)
		go func() {
			hErr <- http.ListenAndServeTLS(addr, certFile, keyFile, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				server.SetQuicHeaders(w.Header())
				s.ServeHTTP(w, r)
			}))
		}()
		go func() {
			qErr <- server.Serve(udpConn)
		}()

		select {
		case err := <-hErr:
			server.Close()
			if err != http.ErrServerClosed {
				panic(err)
			}
		case err := <-qErr:
			// Cannot close the HTTP server or wait for requests to complete properly :/
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	if fn != nil {
		defer fn()
	}
	s.Emit("listening")

	return s
}

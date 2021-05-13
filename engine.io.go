package engineio

import (
	"net/http"
)

const Protocol = 1

func New(arguments ...interface{}) types.Server {
	if len(arguments) > 0 {
		switch s := arguments[0].(type) {
		case *HttpServer:
			if s1, ok := arguments[1]; ok {
				if c, ck := s1.(*types.Config); ck {
					return Attach(s, c)
				}
			}
			return Attach(s, nil)
		case *types.Config:
			return NewServer(s)
		}
	}
	return NewServer(nil)
}

func Listen(addr string, options *types.Config, fn types.Fn) types.Server {
	server := createServer(addr, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "Not Implemented", 501)
	}))

	// create engine server
	engine := Attach(server, options)
	engine.HttpServer(server)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return engine
}

func Attach(server *HttpServer, options *types.Config) types.Server {
	engine := NewServer(options)
	engine.attach(server, options)
	return engine
}

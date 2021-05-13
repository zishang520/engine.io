package engineio

import (
	"net/http"
)

const Protocol = 1

func New(arguments ...interface{}) types.Server {
	if len(arguments) > 0 {
		switch s := arguments[0].(type) {
		case *ServeMux:
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

func Listen(port string, options *types.Config, fn types.Fn) types.Server {
	server := NewServeMux(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(501)
		io.WriteString(w, "Not Implemented")
	}))

	// create engine server
	engine := Attach(server, options)
	engine.HttpServer(server)

	err := http.ListenAndServe(port, server)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return engine
}

func Attach(server *ServeMux, options *types.Config) types.Server {
	engine := NewServer(options)
	engine.attach(server, options)
	return engine
}

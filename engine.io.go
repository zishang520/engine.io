package engineio

import (
	"net/http"
)

const (
	Protocol = 1

	Attach = func(server, options interface{}) {
		engine := NewServer(options)
		engine.attach(server, options)
		return engine
	}

	Listen = func(port, options, fn) {
		http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(501)
			io.WriteString(w, "Not Implemented")
		}))
		server := http.createServer(func(req, res) {
			res.writeHead(501)
			res.end(`Not Implemented`)
		})
		// create engine server
		engine := Attach(server, options)
		engine.httpServer = server

		http.listen(port, fn)

		return engine
	}
)

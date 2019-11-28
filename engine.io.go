package engineio

var (
	Protocol int = 1

	Attach = func(server, options) {
		engine := NewServer(options)
		engine.Attach(server, options)
		return engine
	}

	Listen = func(port, options, fn) {
		server := http.createServer(func(req, res) {
			res.writeHead(501)
			res.end(`Not Implemented`)
		})
		// create engine server
		engine := Attach(server, options)
		engine.httpServer = server

		server.listen(port, fn)

		return engine
	}
)

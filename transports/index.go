package transports

var (
	Polling = func(req) {
		if `string` == req._query.j {
			return NewJSONP(req)
		} else {
			return NewXHR(req)
		}
	}
	UpgradesTo = []string{`websocket`}
)

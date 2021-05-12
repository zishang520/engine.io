package types

import (
	"net/http"
)

type HttpContext struct {
	Request   *http.Request
	Response  http.ResponseWriter
	Websocket Socket

	Cleanup Fn
}

type HttpCompression struct {
	Threshold int
}

type PerMessageDeflate struct {
	Threshold int
}

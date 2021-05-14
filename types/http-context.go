package types

import (
	"context"
	"net/http"
)

type HttpContext struct {
	Request   *http.Request
	Response  http.ResponseWriter
	Websocket *WebSocketConn

	Context context.Context

	Cleanup Fn
}

type HttpCompression struct {
	Threshold int
}

type PerMessageDeflate struct {
	Threshold int
}

package types

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
)

type HttpContext struct {
	Request   *http.Request
	Response  http.ResponseWriter
	Websocket *websocket.Conn

	Context context.Context

	Cleanup Fn
}

type HttpCompression struct {
	Threshold int
}

type PerMessageDeflate struct {
	Threshold int
}

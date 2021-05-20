package types

import (
	"github.com/valyala/fasthttp"
)

type HttpContext struct {
	*fasthttp.RequestCtx
	Websocket *WebSocketConn

	Cleanup Fn
}

type HttpCompression struct {
	Threshold int
}

type PerMessageDeflate struct {
	Threshold int
}

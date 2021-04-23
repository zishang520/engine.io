package transports

import (
	"github.com/zishang520/engine.io/types"
)

type Func func(*types.HttpContext) Transport

var Transports map[string]Func = map[string]Func{
	"polling": func(ctx *types.HttpContext) Transport {
		if _, has := ctx.Request.URL.Query()["j"]; has {
			return NewJSONP(ctx)
		}
		return NewPolling(ctx)
	},
	"websocket": func(*types.HttpContext) Transport {
		return NewWebSocket(ctx)
	},
}

package transports

import (
	"github.com/zishang520/engine.io/types"
)

type Func func(*types.HttpContext) types.Transport

var Transports map[string]Func = map[string]Func{
	"polling": func(ctx *types.HttpContext) types.Transport {
		if _, has := ctx.Request.URL.Query()["j"]; has {
			return NewJSONP(ctx)
		}
		return NewPolling(ctx)
	},
	"websocket": func(ctx *types.HttpContext) types.Transport {
		return NewWebSocket(ctx)
	},
}

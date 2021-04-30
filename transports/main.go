package transports

import (
	"github.com/zishang520/engine.io/types"
)

type Func func(*types.HttpContext) types.Transport

type T struct {
	New             Func
	HandlesUpgrades bool
}

var (
	Transports map[string]*T = map[string]*T{

		"polling": &T{
			New: func(ctx *types.HttpContext) types.Transport {
				if _, has := ctx.Request.URL.Query()["j"]; has {
					return NewJSONP(ctx)
				}
				return NewPolling(ctx)
			},
			HandlesUpgrades: false,
		},

		"websocket": &T{
			New: func(ctx *types.HttpContext) types.Transport {
				return NewWebSocket(ctx)
			},
			HandlesUpgrades: true,
		},
	}
)

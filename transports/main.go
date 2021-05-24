package transports

import (
	"github.com/zishang520/engine.io/types"
)

type Func func(*types.HttpContext) types.Transport

type _local struct {
	New             Func
	HandlesUpgrades bool
	UpgradesTo      func() *types.Set
}

var Transports map[string]*_local = map[string]*_local{
	"polling": &_local{
		New: func(ctx *types.HttpContext) types.Transport {
			if ctx.QueryArgs().Has("j") {
				return NewJSONP(ctx)
			}
			return NewPolling(ctx)
		},
		HandlesUpgrades: false,
		UpgradesTo: func() *types.Set {
			return &types.Set{}
		},
	},

	"websocket": &_local{
		New: func(ctx *types.HttpContext) types.Transport {
			return NewWebSocket(ctx)
		},
		HandlesUpgrades: true,
		UpgradesTo: func() *types.Set {
			return &types.Set{"websocket": types.NULL}
		},
	},
}

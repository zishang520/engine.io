package transports

import (
	"github.com/zishang520/engine.io/types"
)

type transports struct {
	New             func(*types.HttpContext) Transport
	HandlesUpgrades bool
	UpgradesTo      *types.Set
}

var _transports map[string]*transports = map[string]*transports{
	"polling": &transports{
		New: func(ctx *types.HttpContext) Transport {
			if ctx.Query().Has("j") {
				return NewJSONP(ctx)
			}
			return NewPolling(ctx)
		},
		HandlesUpgrades: false,
		UpgradesTo:      types.NewSet("websocket"),
	},

	"websocket": &transports{
		New: func(ctx *types.HttpContext) Transport {
			return NewWebSocket(ctx)
		},
		HandlesUpgrades: true,
		UpgradesTo:      types.NewSet(),
	},
}

func Transports() map[string]*transports {
	return _transports
}

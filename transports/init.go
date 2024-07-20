package transports

import (
	"github.com/zishang520/engine.io/v2/types"
)

type transports struct {
	New             func(*types.HttpContext) Transport
	HandlesUpgrades bool
	UpgradesTo      *types.Set[string]
}

const (
	POLLING      string = "polling"
	WEBSOCKET    string = "websocket"
	WEBTRANSPORT string = "webtransport"
)

var _transports map[string]*transports

func init() {
	_transports = map[string]*transports{
		POLLING: {
			// Polling polymorphic New.
			New: func(ctx *types.HttpContext) Transport {
				if ctx.Query().Has("j") {
					return NewJSONP(ctx)
				}
				return NewPolling(ctx)
			},
			HandlesUpgrades: false,
			UpgradesTo:      types.NewSet(WEBSOCKET, WEBTRANSPORT),
		},

		WEBSOCKET: {
			New: func(ctx *types.HttpContext) Transport {
				return NewWebSocket(ctx)
			},
			HandlesUpgrades: true,
			UpgradesTo:      types.NewSet[string](),
		},

		WEBTRANSPORT: {
			New: func(ctx *types.HttpContext) Transport {
				return NewWebTransport(ctx)
			},
			HandlesUpgrades: true,
			UpgradesTo:      types.NewSet[string](),
		},
	}
}

func Transports() map[string]*transports {
	return _transports
}

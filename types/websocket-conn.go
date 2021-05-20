package types

import (
	"github.com/fasthttp/websocket"
	"github.com/zishang520/engine.io/events"
)

type WebSocketConn struct {
	events.EventEmitter
	*websocket.Conn
}

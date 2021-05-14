package types

import (
	"github.com/fasthttp/websocket"
	"github.com/zishang520/engine.io/events"
	"io"
)

type WebSocketConn struct {
	events.EventEmitter
	*websocket.Conn
}

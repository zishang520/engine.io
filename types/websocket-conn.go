package types

import (
	"github.com/gorilla/websocket"
	"github.com/zishang520/engine.io/v2/events"
)

type WebSocketConn struct {
	events.EventEmitter
	*websocket.Conn
}

func (t *WebSocketConn) Close() error {
	defer t.Emit("close")
	return t.Conn.Close()
}

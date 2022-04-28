package types

import (
	"github.com/gorilla/websocket"
	"github.com/zishang520/engine.io/events"
)

type WebSocketConn struct {
	events.EventEmitter
	*websocket.Conn
}

package types

import (
	"github.com/gorilla/websocket"
	"github.com/zishang520/engine.io/events"
	"io"
)

type WebSocketConn struct {
	events.EventEmitter
	*websocket.Conn
}

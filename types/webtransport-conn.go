package types

import (
	_webtransport "github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io/v2/events"
	"github.com/zishang520/engine.io/v2/webtransport"
)

type WebTransportConn struct {
	events.EventEmitter
	*webtransport.Conn
}

func (t *WebTransportConn) CloseWithError(code _webtransport.SessionErrorCode, msg string) error {
	defer t.Emit("close")
	return t.Conn.CloseWithError(code, msg)
}

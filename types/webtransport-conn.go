package types

import (
	wt "github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io/v2/webtransport"
)

type WebTransportConn struct {
	EventEmitter

	*webtransport.Conn
}

func (t *WebTransportConn) CloseWithError(code wt.SessionErrorCode, msg string) error {
	defer t.Emit("close")
	return t.Conn.CloseWithError(code, msg)
}

package types

import (
	"github.com/quic-go/webtransport-go"
	"github.com/zishang520/engine.io/events"
)

type WebTransportConn struct {
	events.EventEmitter
	*webtransport.Session

	Stream webtransport.Stream
}

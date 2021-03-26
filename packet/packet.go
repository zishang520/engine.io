package packet

import (
	"io"
)

type Type string

const (
	OPEN    Type = "open"
	CLOSE   Type = "close"
	PING    Type = "ping"
	PONG    Type = "pong"
	MESSAGE Type = "message"
	UPGRADE Type = "upgrade"
	NOOP    Type = "noop"
	ERROR   Type = "error"
)

type Packet struct {
	Type Type      `json:"type"`
	Data io.Reader `json:"data"`
}

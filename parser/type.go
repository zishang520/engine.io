package parser

import (
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"io"
)

type Paser interface {
	Protocol() int
	EncodePacket(*packet.Packet, bool) (*types.BytesBuffer, error)
	DecodePacket(io.Reader) (*packet.Packet, error)
	EncodePayload([]*packet.Packet, ...bool) (*types.BytesBuffer, error)
	DecodePayload(io.Reader) []*packet.Packet
}

/**
 * Packet types.
 */
var (
	PACKET_TYPES map[packet.Type]byte = map[packet.Type]byte{
		packet.OPEN:    '0',
		packet.CLOSE:   '1',
		packet.PING:    '2',
		packet.PONG:    '3',
		packet.MESSAGE: '4',
		packet.UPGRADE: '5',
		packet.NOOP:    '6',
	}

	PACKET_TYPES_REVERSE map[byte]packet.Type = map[byte]packet.Type{
		'0': packet.OPEN,
		'1': packet.CLOSE,
		'2': packet.PING,
		'3': packet.PONG,
		'4': packet.MESSAGE,
		'5': packet.UPGRADE,
		'6': packet.NOOP,
	}

	ERROR_PACKET = &packet.Packet{Type: packet.ERROR, Data: types.NewStringBufferString(`parser error`)}

	SEPARATOR = byte(0x1E)
)

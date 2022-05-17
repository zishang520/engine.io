package parser

import (
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
)

type Parser interface {
	Protocol() int
	EncodePacket(*packet.Packet, bool, ...bool) (types.BufferInterface, error)
	DecodePacket(types.BufferInterface, ...bool) (*packet.Packet, error)
	EncodePayload([]*packet.Packet, ...bool) (types.BufferInterface, error)
	DecodePayload(types.BufferInterface) []*packet.Packet
}

// Packet types.
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

	// Premade error packet.
	ERROR_PACKET = &packet.Packet{Type: packet.ERROR, Data: types.NewStringBufferString(`parser error`)}
)

const SEPARATOR = byte(0x1E)

package parser

import (
	"encoding/base64"
	"fmt"
	"github.com/zishang520/engine.io/bytes"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/types"
	"io"
)

/**
 * Current protocol version.
 */
const Protocol = 4

/**
 * Packet types.
 */
var (
	PACKET_TYPES map[string]byte = map[string]byte{
		"open":    '0',
		"close":   '1',
		"ping":    '2',
		"pong":    '3',
		"message": '4',
		"upgrade": '5',
		"noop":    '6',
	}

	PACKET_TYPES_REVERSE map[byte]string = map[byte]string{
		'0': "open",
		'1': "close",
		'2': "ping",
		'3': "pong",
		'4': "message",
		'5': "upgrade",
		'6': "noop",
	}

	ERROR_PACKET = &types.Packet{Type: `error`, Data: bytes.NewStringBuffer([]byte(`parser error`))}

	SEPARATOR = byte(0x1E)
)

func EncodePacket(packet *types.Packet, supportsBinary bool) (*bytes.Buffer, error) {
	encode := bytes.NewBuffer(nil)

	if packet == nil {
		return encode, errors.New(`Packet is nil`)
	}

	if c, ok := packet.Data.(io.Closer); ok {
		defer c.Close()
	}

	_type, _type_ok := PACKET_TYPES[packet.Type]
	if !_type_ok {
		return encode, errors.New(`Packet Type error`)
	}

	switch v := packet.Data.(type) {
	case *bytes.StringBuffer:
		encode.WriteByte(_type)
		v.WriteTo(encode)
	case *bytes.Buffer:
		if !supportsBinary {
			encode.WriteByte('b')
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			v.WriteTo(b64)
		} else {
			v.WriteTo(encode)
		}
	default:
		encode.WriteByte(_type)
	}
	return encode, nil
}

func DecodePacket(data io.Reader) (*types.Packet, error) {
	if data == nil {
		return ERROR_PACKET, errors.New(`parser error`)
	}

	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	switch v := data.(type) {
	case *bytes.StringBuffer:
		msgType, err := v.ReadByte()
		if err != nil {
			return ERROR_PACKET, err
		}
		if msgType == 'b' {
			decode := bytes.NewBuffer(nil)
			decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v))
			return &types.Packet{
				Type: "message",
				Data: decode,
			}, nil
		}
		packetType, ok := PACKET_TYPES_REVERSE[msgType]
		if !ok {
			return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType))
		}
		stringBuffer := bytes.NewStringBuffer(nil)
		stringBuffer.ReadFrom(v)
		return &types.Packet{
			Type: packetType,
			Data: stringBuffer,
		}, nil
	default:
		decode := bytes.NewBuffer(nil)
		decode.ReadFrom(v)
		return &types.Packet{
			Type: "message",
			Data: decode,
		}, nil
	}

	return ERROR_PACKET, errors.New(`parser error`)
}

func EncodePayload(packets []*types.Packet) (*bytes.Buffer, error) {
	enPayload := bytes.NewBuffer(nil)

	for _, packet := range packets {
		if buf, err := EncodePacket(packet, false); err != nil {
			return enPayload, err
		} else {
			if enPayload.Len() > 0 {
				enPayload.WriteByte(SEPARATOR)
			}
			buf.WriteTo(enPayload)
		}
	}

	return enPayload, nil
}

func DecodePayload(data *bytes.Buffer) []*types.Packet {
	packets := []*types.Packet{}

	for buf, err := data.ReadBytes(SEPARATOR); err != io.EOF; {
		if packet, err := DecodePacket(bytes.NewBuffer(buf)); err == nil {
			packets = append(packets, packet)
		} else {
			break
		}
	}

	return packets
}

package parser

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"io"
)

/**
 * Current protocol version.
 */
var ParserV4 parserv4

type parserv4 struct{}

func (parserv4) Protocol() int {
	return 4
}

func (p parserv4) EncodePacket(packet *packet.Packet, supportsBinary bool) (*types.BytesBuffer, error) {
	encode := types.NewBytesBuffer(nil)

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
	case *types.StringBuffer:
		encode.WriteByte(_type)
		v.WriteTo(encode)
	case io.Reader, io.Writer:
		if !supportsBinary {
			encode.WriteByte('b')
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			io.Copy(b64, v)
		} else {
			io.Copy(encode, v)
		}
	default:
		encode.WriteByte(_type)
	}
	return encode, nil
}

func (p parserv4) DecodePacket(data io.Reader) (*packet.Packet, error) {
	if data == nil {
		return ERROR_PACKET, errors.New(`parser error`)
	}

	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	switch v := data.(type) {
	case *types.StringBuffer:
		msgType, err := v.ReadByte()
		if err != nil {
			return ERROR_PACKET, err
		}
		if msgType == 'b' {
			decode := types.NewBytesBuffer(nil)
			decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v))
			return &packet.Packet{
				Type: packet.MESSAGE,
				Data: decode,
			}, nil
		}
		packetType, ok := PACKET_TYPES_REVERSE[msgType]
		if !ok {
			return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType))
		}
		stringBuffer := types.NewStringBuffer(nil)
		stringBuffer.ReadFrom(v)
		return &packet.Packet{
			Type: packetType,
			Data: stringBuffer,
		}, nil
	default:
		decode := types.NewBytesBuffer(nil)
		io.Copy(decode, v)
		return &packet.Packet{
			Type: packet.MESSAGE,
			Data: decode,
		}, nil
	}

	return ERROR_PACKET, errors.New(`parser error`)
}

func (p parserv4) EncodePayload(packets []*packet.Packet) (*types.BytesBuffer, error) {
	enPayload := types.NewBytesBuffer(nil)

	for _, packet := range packets {
		if buf, err := p.EncodePacket(packet, false); err != nil {
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

func (p parserv4) DecodePayload(data *types.BytesBuffer) []*packet.Packet {
	packets := []*packet.Packet{}

	for {
		buf, err := data.ReadBytes(SEPARATOR)
		if packet, err := p.DecodePacket(types.NewStringBuffer(bytes.TrimSuffix(buf, []byte{SEPARATOR}))); err == nil {
			packets = append(packets, packet)
		} else {
			break
		}
		if err == io.EOF {
			break
		}
	}

	return packets
}

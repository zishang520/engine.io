package parser

import (
	"bufio"
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
var ParserV4 Paser = parserv4{}

type parserv4 struct{}

func (parserv4) Protocol() int {
	return 4
}

func (p parserv4) EncodePacket(packet *packet.Packet, supportsBinary bool, _ ...bool) (types.PacketBuffer, error) {
	if packet == nil {
		return nil, errors.New(`Packet is nil`)
	}

	if c, ok := packet.Data.(io.Closer); ok {
		defer c.Close()
	}

	_type, _type_ok := PACKET_TYPES[packet.Type]
	if !_type_ok {
		return nil, errors.New(`Packet Type error`)
	}

	switch v := packet.Data.(type) {
	case *types.StringBuffer:
		encode := types.NewStringBuffer(nil)
		encode.WriteByte(_type)
		v.WriteTo(encode)

		return encode, nil
	case io.Reader:
		if !supportsBinary {
			encode := types.NewStringBuffer(nil)
			encode.WriteByte('b')
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			io.Copy(b64, v)
			return encode, nil
		}
		encode := types.NewBytesBuffer(nil)
		io.Copy(encode, v)
		return encode, nil
	}
	encode := types.NewStringBuffer(nil)
	encode.WriteByte(_type)
	return encode, nil
}

func (p parserv4) DecodePacket(data io.Reader, _ ...bool) (*packet.Packet, error) {
	if data == nil {
		return ERROR_PACKET, errors.New(`parser error`)
	}

	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	// 字符串
	if v, ok := data.(*types.StringBuffer); ok {
		msgType, err := v.ReadByte()
		if err != nil {
			return ERROR_PACKET, err
		}
		if msgType == 'b' {
			decode := types.NewBytesBuffer(nil)
			decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v))
			return &packet.Packet{Type: packet.MESSAGE, Data: decode}, nil
		}
		packetType, ok := PACKET_TYPES_REVERSE[msgType]
		if !ok {
			return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType))
		}
		stringBuffer := types.NewStringBuffer(nil)
		stringBuffer.ReadFrom(v)
		return &packet.Packet{Type: packetType, Data: stringBuffer}, nil
	}

	// 二进制
	decode := types.NewBytesBuffer(nil)
	io.Copy(decode, data)
	return &packet.Packet{Type: packet.MESSAGE, Data: decode}, nil
}

func (p parserv4) EncodePayload(packets []*packet.Packet, _ ...bool) (types.PacketBuffer, error) {
	enPayload := types.NewStringBuffer(nil)

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

func (p parserv4) DecodePayload(data io.Reader) (packets []*packet.Packet) {
	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	scanner := bufio.NewScanner(data)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, SEPARATOR); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	for scanner.Scan() {
		if packet, err := p.DecodePacket(types.NewStringBuffer(scanner.Bytes())); err == nil {
			packets = append(packets, packet)
		}
	}

	return packets
}

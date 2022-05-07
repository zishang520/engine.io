package parser

import (
	"encoding/base64"
	"fmt"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type parserv3 struct{}

var (
	defaultParserv3 Parser = &parserv3{}
)

func Parserv3() Parser {
	return defaultParserv3
}

// Current protocol version.
func (*parserv3) Protocol() int {
	return 3
}

func (p *parserv3) EncodePacket(data *packet.Packet, supportsBinary bool, utf8encode ...bool) (types.BufferInterface, error) {
	utf8encode = append(utf8encode, false)

	if c, ok := data.Data.(io.Closer); ok {
		defer c.Close()
	}

	switch v := data.Data.(type) {
	case *types.StringBuffer, *strings.Reader:
		encode := types.NewStringBuffer(nil)
		if err := encode.WriteByte(PACKET_TYPES[data.Type]); err != nil {
			return nil, err
		}
		if utf8encode[0] {
			if _, err := io.Copy(utils.NewUtf8Encoder(encode), v); err != nil {
				return nil, err
			}
		} else {
			if _, err := io.Copy(encode, v); err != nil {
				return nil, err
			}
		}
		return encode, nil
	case io.Reader:
		if !supportsBinary {
			encode := types.NewStringBuffer(nil)
			if _, err := encode.Write([]byte{'b', PACKET_TYPES[data.Type]}); err != nil {
				return nil, err
			}
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			if _, err := io.Copy(b64, v); err != nil {
				return nil, err
			}
			return encode, nil
		}
		encode := types.NewBytesBuffer(nil)
		if err := encode.WriteByte(PACKET_TYPES[data.Type] - '0'); err != nil {
			return nil, err
		}
		if _, err := io.Copy(encode, v); err != nil {
			return nil, err
		}
		return encode, nil
	}
	// default nil
	encode := types.NewStringBuffer(nil)
	if err := encode.WriteByte(PACKET_TYPES[data.Type]); err != nil {
		return nil, err
	}
	return encode, nil
}

// Decodes a packet. Data also available as an ArrayBuffer if requested.
func (p *parserv3) DecodePacket(data types.BufferInterface, utf8decode ...bool) (*packet.Packet, error) {
	utf8decode = append(utf8decode, false)
	if data == nil {
		return ERROR_PACKET, errors.New(`parser error`).Err()
	}

	msgType, err := data.ReadByte()
	if err != nil {
		return ERROR_PACKET, err
	}

	switch v := data.(type) {
	case *types.StringBuffer:
		if msgType == 'b' {
			msgType, err = data.ReadByte()
			if err != nil {
				return ERROR_PACKET, err
			}
			packetType, ok := PACKET_TYPES_REVERSE[msgType]
			if !ok {
				return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType)).Err()
			}
			decode := types.NewBytesBuffer(nil)
			if _, err := decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v)); err != nil {
				return ERROR_PACKET, err
			}
			return &packet.Packet{Type: packetType, Data: decode}, nil
		}
		packetType, ok := PACKET_TYPES_REVERSE[msgType]
		if !ok {
			return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType)).Err()
		}
		decode := types.NewStringBuffer(nil)
		if utf8decode[0] {
			if _, err := decode.ReadFrom(utils.NewUtf8Decoder(v)); err != nil {
				return ERROR_PACKET, err
			}
		} else {
			if _, err := decode.ReadFrom(v); err != nil {
				return ERROR_PACKET, err
			}
		}
		return &packet.Packet{Type: packetType, Data: decode}, nil
	}

	// Default
	packetType, ok := PACKET_TYPES_REVERSE[msgType+'0']
	if !ok {
		return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType+'0')).Err()
	}
	decode := types.NewBytesBuffer(nil)
	if _, err := io.Copy(decode, data); err != nil {
		return ERROR_PACKET, err
	}
	return &packet.Packet{Type: packetType, Data: decode}, nil
}

func (p *parserv3) hasBinary(packets []*packet.Packet) bool {
	if len(packets) == 0 {
		return false
	}
	for _, packet := range packets {
		switch packet.Data.(type) {
		case *types.StringBuffer:
		case *strings.Reader:
		case nil:
		default:
			return true
		}
	}

	return false
}

func (p *parserv3) EncodePayload(packets []*packet.Packet, supportsBinary ...bool) (types.BufferInterface, error) {
	supportsBinary = append(supportsBinary, false)

	if supportsBinary[0] && p.hasBinary(packets) {
		return p.EncodePayloadAsBinary(packets)
	}

	enPayload := types.NewStringBuffer(nil)

	if len(packets) == 0 {
		if _, err := enPayload.WriteString(`0:`); err != nil {
			return nil, err
		}
		return enPayload, nil
	}

	for _, packet := range packets {
		buf, err := p.EncodePacket(packet, supportsBinary[0], false)
		if err != nil {
			return nil, err
		}
		if _, err := enPayload.WriteString(fmt.Sprintf(`%d:%s`, utils.Utf16Count(buf.Bytes()), buf.String())); err != nil {
			return nil, err
		}
	}

	return enPayload, nil
}

func (p *parserv3) encodeOneBinaryPacket(packet *packet.Packet) (types.BufferInterface, error) {
	binarypacket := types.NewBytesBuffer(nil)

	buf, err := p.EncodePacket(packet, true, true)
	if err != nil {
		return nil, err
	}

	if _, ok := buf.(*types.StringBuffer); ok {
		encodingLength := fmt.Sprintf(`%d`, utils.Utf16Count(buf.Bytes())) // JS length
		if err := binarypacket.WriteByte(0); err != nil {
			return nil, err
		}
		for i, l := 0, len(encodingLength); i < l; i++ {
			if err := binarypacket.WriteByte(encodingLength[i] - '0'); err != nil {
				return nil, err
			}
		}
		if err := binarypacket.WriteByte(0xFF); err != nil {
			return nil, err
		}
		if _, err := buf.WriteTo(utils.NewUtf8Encoder(binarypacket)); err != nil {
			return nil, err
		}

		return binarypacket, nil
	}

	encodingLength := fmt.Sprintf(`%d`, buf.Len())
	// is binary (true binary = 1)
	if err := binarypacket.WriteByte(1); err != nil {
		return nil, err
	}
	for i, l := 0, len(encodingLength); i < l; i++ {
		if err := binarypacket.WriteByte(encodingLength[i] - '0'); err != nil {
			return nil, err
		}
	}
	if err := binarypacket.WriteByte(0xFF); err != nil {
		return nil, err
	}
	if _, err := binarypacket.ReadFrom(buf); err != nil {
		return nil, err
	}

	return binarypacket, nil
}

func (p *parserv3) EncodePayloadAsBinary(packets []*packet.Packet) (types.BufferInterface, error) {
	enPayload := types.NewBytesBuffer(nil)

	if len(packets) == 0 {
		return enPayload, nil
	}

	for _, packet := range packets {
		buf, err := p.encodeOneBinaryPacket(packet)
		if err != nil {
			return nil, err
		}
		if _, err := enPayload.ReadFrom(buf); err != nil {
			return nil, err
		}
	}

	return enPayload, nil
}

func (p *parserv3) DecodePayload(data types.BufferInterface) (packets []*packet.Packet) {
	switch v := data.(type) {
	case *types.StringBuffer:
		PACKETLEN := 0
		for v.Len() > 0 {
			length, err := v.ReadString(':')
			if err != nil {
				return packets
			}
			_l := len(length)
			if _l < 1 {
				return packets
			}
			packetLen, err := strconv.ParseInt(length[:_l-1], 10, 0)
			if err != nil {
				return packets
			}

			PACKETLEN = int(packetLen)
			msg := types.NewStringBuffer(nil)
			for i := 0; i < PACKETLEN; {
				r, _, e := v.ReadRune()
				if e != nil {
					return packets
				}
				i += utils.Utf16Len(r)
				if _, err := msg.WriteRune(r); err != nil {
					return packets
				}
			}

			if msg.Len() > 0 {
				packet, err := p.DecodePacket(msg, false)
				if err != nil {
					// parser error in individual packet - ignoring payload
					return packets
				}
				packets = append(packets, packet)
			}
		}
		return packets
	}
	return p.DecodePayloadAsBinary(data)
}

func (p *parserv3) DecodePayloadAsBinary(bufferTail types.BufferInterface) (packets []*packet.Packet) {
	PACKETLEN := 0
	for bufferTail.Len() > 0 {
		startByte, err := bufferTail.ReadByte()
		if err != nil {
			// parser error in individual packet - ignoring payload
			return packets
		}
		isString := startByte == 0x00
		length, err := bufferTail.ReadBytes(0xFF)
		if err != nil {
			return packets
		}
		_l := len(length)
		if _l < 1 {
			return packets
		}
		lenByte := length[:_l-1]
		for k, l := 0, len(lenByte); k < l; k++ {
			lenByte[k] = lenByte[k] + '0'
		}
		packetLen, err := strconv.ParseInt(string(lenByte), 10, 0)
		if err != nil {
			return packets
		}
		PACKETLEN = int(packetLen)
		if isString {
			data := types.NewStringBuffer(nil)
			buf := make([]byte, 0, 4) // rune bytes
			for k := 0; k < PACKETLEN; {
				// read utf8
				for len(buf) < 4 {
					r, _, err := bufferTail.ReadRune()
					if err != nil {
						if err == io.EOF {
							break
						}
						return packets
					}
					if !utf8.ValidRune(r) {
						r = 0xFFFD
					}
					buf = append(buf, byte(r))
				}
				r, l := utf8.DecodeRune(buf)
				k += utils.Utf16Len(r)
				if _, err := data.Write(utils.Utf8decodeBytes(buf[0:l])); err != nil {
					return packets
				}
				buf = buf[l:]
			}
			if cursor := len(utils.Utf8encodeBytes(buf)); cursor > 0 {
				bufferTail.Seek(-int64(cursor), io.SeekCurrent)
			}
			if data.Len() > 0 {
				packet, err := p.DecodePacket(data, false)
				if err != nil {
					// parser error in individual packet - ignoring payload
					return packets
				}
				packets = append(packets, packet)
			}
		} else {
			if data := bufferTail.Next(PACKETLEN); len(data) > 0 {
				packet, err := p.DecodePacket(types.NewBytesBuffer(data), false)
				if err != nil {
					// parser error in individual packet - ignoring payload
					return packets
				}
				packets = append(packets, packet)
			}
		}
	}

	return packets
}
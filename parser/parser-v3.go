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
	"unicode/utf8"
)

/**
 * Current protocol version.
 */
var ParserV3 Paser = parserv3{}

type parserv3 struct{}

func (parserv3) Protocol() int {
	return 3
}

func (p parserv3) EncodePacket(packet *packet.Packet, supportsBinary bool, utf8encode ...bool) (types.PacketBuffer, error) {
	utf8encode = append(utf8encode, false)

	if c, ok := packet.Data.(io.Closer); ok {
		defer c.Close()
	}

	switch v := packet.Data.(type) {
	case *types.StringBuffer:
		encode := types.NewStringBuffer(nil)
		encode.WriteByte(PACKET_TYPES[packet.Type])
		if utf8encode[0] {
			v.WriteTo(utils.NewUtf8Encoder(encode))
		} else {
			v.WriteTo(encode)
		}
		return encode, nil
	case io.Reader:
		if !supportsBinary {
			encode := types.NewStringBuffer(nil)
			encode.Write([]byte{'b', PACKET_TYPES[packet.Type]})
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			io.Copy(b64, v)
			return encode, nil
		}
		encode := types.NewBytesBuffer(nil)
		encode.WriteByte(PACKET_TYPES[packet.Type] - '0')
		io.Copy(encode, v)
		return encode, nil
	}
	encode := types.NewStringBuffer(nil)
	encode.WriteByte(PACKET_TYPES[packet.Type])
	return encode, nil
}

/**
 * Decodes a packet. Data also available as an ArrayBuffer if requested.
 *
 * @return {Object} with `type` and `data` (if any)
 * @api public
 */

func (p parserv3) DecodePacket(data io.Reader, utf8decode ...bool) (*packet.Packet, error) {
	utf8decode = append(utf8decode, false)
	if data == nil {
		return ERROR_PACKET, errors.New(`parser error`)
	}
	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	msgType := []byte{0xFF}
	if _, err := data.Read(msgType); err != nil {
		return ERROR_PACKET, err
	}

	if v, ok := data.(*types.StringBuffer); ok {
		if msgType[0] == 'b' {
			if _, err := data.Read(msgType); err != nil {
				return ERROR_PACKET, err
			}
			packetType, ok := PACKET_TYPES_REVERSE[msgType[0]]
			if !ok {
				return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
			}
			decode := types.NewBytesBuffer(nil)
			decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v))
			return &packet.Packet{Type: packetType, Data: decode}, nil
		}
		packetType, ok := PACKET_TYPES_REVERSE[msgType[0]]
		if !ok {
			return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
		}
		decode := types.NewStringBuffer(nil)
		if utf8decode[0] {
			decode.ReadFrom(utils.NewUtf8Decoder(v))
		} else {
			decode.ReadFrom(v)
		}
		return &packet.Packet{Type: packetType, Data: decode}, nil
	}

	packetType, ok := PACKET_TYPES_REVERSE[msgType[0]+'0']
	if !ok {
		return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]+'0'))
	}
	decode := types.NewBytesBuffer(nil)
	io.Copy(decode, data)
	return &packet.Packet{Type: packetType, Data: decode}, nil
}

func (p parserv3) hasBinary(packets []*packet.Packet) bool {
	if len(packets) == 0 {
		return false
	}
	for _, packet := range packets {
		switch packet.Data.(type) {
		case *types.StringBuffer:
			break
		case io.Reader:
			return true
			break
		case nil:
			break
		default:
			break
		}
	}

	return false
}

func (p parserv3) EncodePayload(packets []*packet.Packet, supportsBinary ...bool) (types.PacketBuffer, error) {
	supportsBinary = append(supportsBinary, false)

	if supportsBinary[0] && p.hasBinary(packets) {
		return p.EncodePayloadAsBinary(packets)
	}

	enPayload := types.NewStringBuffer(nil)

	if len(packets) == 0 {
		enPayload.WriteString(`0:`)
		return enPayload, nil
	}

	for _, packet := range packets {
		buf, err := p.EncodePacket(packet, supportsBinary[0], false)
		if err != nil {
			return enPayload, err
		}
		enPayload.WriteString(fmt.Sprintf(`%d:%s`, utils.Utf16Count(buf.Bytes()), buf.String()))
	}

	return enPayload, nil
}

func (p parserv3) encodeOneBinaryPacket(packet *packet.Packet) (types.PacketBuffer, error) {
	binarypacket := types.NewBytesBuffer(nil)

	buf, err := p.EncodePacket(packet, true, true)
	if err != nil {
		return binarypacket, err
	}

	if _, ok := buf.(*types.StringBuffer); ok {
		encodingLength := fmt.Sprintf(`%d`, utils.Utf16Count(buf.Bytes())) // JS length
		binarypacket.WriteByte(0)
		for i := 0; i < len(encodingLength); i++ {
			binarypacket.WriteByte(encodingLength[i] - '0')
		}
		binarypacket.WriteByte(0xFF)
		buf.WriteTo(utils.NewUtf8Encoder(binarypacket))

		return binarypacket, nil
	}

	encodingLength := fmt.Sprintf(`%d`, buf.Len())
	binarypacket.WriteByte(1) // is binary (true binary = 1)
	for i := 0; i < len(encodingLength); i++ {
		binarypacket.WriteByte(encodingLength[i] - '0')
	}
	binarypacket.WriteByte(0xFF)
	binarypacket.ReadFrom(buf)

	return binarypacket, nil
}

func (p parserv3) EncodePayloadAsBinary(packets []*packet.Packet) (types.PacketBuffer, error) {
	enPayload := types.NewBytesBuffer(nil)

	if len(packets) == 0 {
		return enPayload, nil
	}

	for _, packet := range packets {
		buf, err := p.encodeOneBinaryPacket(packet)
		if err != nil {
			return enPayload, err
		}
		enPayload.ReadFrom(buf)
	}

	return enPayload, nil
}

func (p parserv3) DecodePayload(data io.Reader) (packets []*packet.Packet) {
	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	if v, ok := data.(*types.StringBuffer); ok {
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
			packetLen, err := strconv.ParseInt(length[:_l-1], 10, 64)
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
				msg.WriteRune(r)
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

func (p parserv3) DecodePayloadAsBinary(data io.Reader) (packets []*packet.Packet) {
	bufferTail := types.NewBuffer(nil)
	bufferTail.ReadFrom(data)

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
		packetLen, err := strconv.ParseInt(string(lenByte), 10, 64)
		if err != nil {
			return packets
		}
		PACKETLEN = int(packetLen)
		if isString {
			data := types.NewStringBuffer(nil)
			buf := []byte{}
			for k := 0; k < PACKETLEN; {
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
				data.Write(utils.Utf8decodeBytesReturn(buf[0:l]))
				buf = buf[l:]
			}
			if cursor := len(utils.Utf8encodeBytesReturn(buf)); cursor > 0 {
				// Rollback read pointer
				bufferTail.Prev(cursor)
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

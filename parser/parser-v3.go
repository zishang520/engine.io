package parser

import (
	"encoding/base64"
	"fmt"
	"github.com/zishang520/engine.io/bytes"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"io"
	"strconv"
	"strings"
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

func (p parserv3) EncodePacket(packet *packet.Packet, supportsBinary bool, utf8encode bool) (*types.BytesBuffer, error) {
	encode := types.NewBytesBuffer(nil)

	if c, ok := packet.Data.(io.Closer); ok {
		defer c.Close()
	}

	switch v := packet.Data.(type) {
	case *types.StringBuffer:
		encode.WriteByte(PACKET_TYPES[packet.Type])
		if utf8encode {
			v.WriteTo(utils.NewUtf8Encoder(encode))
		} else {
			v.WriteTo(encode)
		}
	case io.Reader:
		if !supportsBinary {
			encode.Write([]byte{'b', PACKET_TYPES[packet.Type]})
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			io.Copy(b64, v)
		} else {
			encode.WriteByte(PACKET_TYPES[packet.Type] - '0')
			io.Copy(encode, v)
		}
	default:
		encode.WriteByte(PACKET_TYPES[packet.Type])
	}
	return encode, nil
}

/**
 * Decodes a packet. Data also available as an ArrayBuffer if requested.
 *
 * @return {Object} with `type` and `data` (if any)
 * @api public
 */

func (p parserv3) DecodePacket(data io.Reader, utf8decode bool) (*packet.Packet, error) {
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

	switch v := data.(type) {
	case *types.StringBuffer:
		if msgType[0] == 'b' {
			if _, err := v.Read(msgType); err != nil {
				return ERROR_PACKET, err
			}
			packetType, ok := PACKET_TYPES[msgType[0]]
			if !ok {
				return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
			}
			decode := types.NewBytesBuffer(nil)
			decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v))
			return &types.Packet{
				Type: packetType,
				Data: decode,
			}, nil
		}
		packetType, ok := PACKET_TYPES[msgType[0]]
		if !ok {
			return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
		}
		stringBuffer := types.NewStringBuffer(nil)
		if utf8decode {
			stringBuffer.ReadFrom(utils.NewUtf8Decoder(v))
		} else {
			stringBuffer.ReadFrom(v)
		}
		return &types.Packet{
			Type: packetType,
			Data: stringBuffer,
		}, nil
	default:
		packetType, ok := PACKET_TYPES[msgType[0]+'0']
		if !ok {
			return ERROR_PACKET, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]+'0'))
		}
		decode := types.NewBytesBuffer(nil)
		io.Copy(decode, v)
		return &types.Packet{
			Type: packetType,
			Data: decode,
		}, nil
	}

	return ERROR_PACKET, errors.New(`parser error`)
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

func (p parserv3) EncodePayload(packets []*packet.Packet, supportsBinary bool) (*types.BytesBuffer, error) {
	isBinary := p.hasBinary(packets)
	if supportsBinary && isBinary {
		return p.EncodePayloadAsBinary(packets)
	}

	enPayload := types.NewBytesBuffer(nil)

	if len(packets) == 0 {
		enPayload.WriteString(`0:`)
		return enPayload, nil
	}

	if !isBinary {
		supportsBinary = false
	}
	for _, packet := range packets {
		buf, err := p.EncodePacket(packet, supportsBinary, false)
		if err != nil {
			return enPayload, err
		}
		enPayload.WriteString(fmt.Sprintf(`%d:%s`, utils.Utf16Count(buf.Bytes()), buf.String()))
	}

	return enPayload, nil
}

func (p parserv3) encodeOneBinaryPacket(packet *packet.Packet) (*bytes.Buffer, error) {
	binarypacket := bytes.NewBuffer(nil)

	buf, err := EncodePacket(packet, true, true)
	if err != nil {
		return binarypacket, err
	}
	switch packet.Data.(type) {
	case *strings.Reader:
		encodingLength := fmt.Sprintf(`%d`, Utf16Count(buf.Bytes())) // JS length
		binarypacket.WriteByte(0)
		for i := 0; i < len(encodingLength); i++ {
			binarypacket.WriteByte(encodingLength[i] - '0')
		}
		binarypacket.WriteByte(0xFF)
		buf.WriteTo(utils.NewUtf8Encoder(binarypacket))
	default:
		encodingLength := fmt.Sprintf(`%d`, buf.Len())
		binarypacket.WriteByte(1) // is binary (true binary = 1)
		for i := 0; i < len(encodingLength); i++ {
			binarypacket.WriteByte(encodingLength[i] - '0')
		}
		binarypacket.WriteByte(0xFF)
		binarypacket.ReadFrom(buf)
	}

	return binarypacket, nil
}

func (p parserv3) EncodePayloadAsBinary(packets []*packet.Packet) (*bytes.Buffer, error) {
	enPayload := bytes.NewBuffer(nil)

	if len(packets) == 0 {
		return enPayload, nil
	}

	for _, packet := range packets {
		if buf, err := encodeOneBinaryPacket(packet); err != nil {
			return enPayload, err
		} else {
			enPayload.ReadFrom(buf)
		}
	}

	return enPayload, nil
}

func (p parserv3) DecodePayload(data io.Reader, callback Callback) bool {
	switch v := data.(type) {
	case *strings.Reader:
		str := bytes.NewBuffer(nil)
		v.WriteTo(str)
		for n, l := 0, Utf16Count(str.Bytes()); str.Len() > 0; {
			length, err := str.ReadString(':')
			if err != nil {
				return callback(ERROR_PACKET, 0, 1)
			}
			_l := len(length)
			if _l < 1 {
				return callback(ERROR_PACKET, 0, 1)
			}
			packetLen, err := strconv.ParseInt(length[:_l-1], 10, 64)
			if err != nil {
				return callback(ERROR_PACKET, 0, 1)
			}

			PACKETLEN := int(packetLen)
			msg := new(strings.Builder)
			for i := 0; i < PACKETLEN; {
				r, _, e := str.ReadRune()
				if e != nil {
					return callback(ERROR_PACKET, 0, 1)
				}
				i += Utf16Len(r)
				msg.WriteRune(r)
			}

			if msg.Len() > 0 {
				packet, err := DecodePacket(strings.NewReader(msg.String()), false)
				if err != nil {
					// parser error in individual packet - ignoring payload
					return callback(ERROR_PACKET, 0, 1)
				}
				if more := callback(packet, n+PACKETLEN+_l-1, l); false == more {
					return more
				}
			}

			n += PACKETLEN + _l
		}
	default:
		return DecodePayloadAsBinary(data, callback)
	}

	return true
}

func (p parserv3) DecodePayloadAsBinary(data io.Reader, callback Callback) bool {
	bufferTail := bytes.NewBuffer(nil)
	bufferTail.ReadFrom(data)

	buffers := []io.Reader{}
	for bufferTail.Len() > 0 {
		startByte, err := bufferTail.ReadByte()
		if err != nil {
			// parser error in individual packet - ignoring payload
			return callback(ERROR_PACKET, 0, 1)
		}
		isString := startByte == 0x00
		length, err := bufferTail.ReadBytes(0xFF)
		if err != nil {
			return callback(ERROR_PACKET, 0, 1)
		}
		_l := len(length)
		if _l < 1 {
			return callback(ERROR_PACKET, 0, 1)
		}
		lenByte := length[:_l-1]
		for k, l := 0, len(lenByte); k < l; k++ {
			lenByte[k] = lenByte[k] + '0'
		}
		packetLen, err := strconv.ParseInt(string(lenByte), 10, 64)
		if err != nil {
			return callback(ERROR_PACKET, 0, 1)
		}
		PACKETLEN := int(packetLen)
		if isString {
			data := new(strings.Builder)
			buf := []byte{}
			for k := 0; k < PACKETLEN; {
				for len(buf) < 4 {
					r, _, err := bufferTail.ReadRune()
					if err != nil {
						if err == io.EOF {
							break
						}
						return callback(ERROR_PACKET, 0, 1)
					}
					if !utf8.ValidRune(r) {
						r = 0xFFFD
					}
					buf = append(buf, byte(r))
				}
				r, l := utf8.DecodeRune(buf)
				k += Utf16Len(r)
				data.Write(Utf8decodeBytesReturn(buf[0:l]))
				buf = buf[l:]
			}
			if cursor := len(Utf8encodeBytesReturn(buf)); cursor > 0 {
				// Rollback read pointer
				bufferTail.Prev(cursor)
			}
			if data.Len() > 0 {
				buffers = append(buffers, strings.NewReader(data.String()))
			}
		} else {
			if data := bufferTail.Next(PACKETLEN); len(data) > 0 {
				buffers = append(buffers, bytes.NewBuffer(data))
			}
		}
	}

	for k, bl := 0, len(buffers); k < bl; k++ {
		packet, err := DecodePacket(buffers[k], false)
		if err != nil {
			// parser error in individual packet - ignoring payload
			return callback(ERROR_PACKET, 0, 1)
		}
		if more := callback(packet, k, bl); false == more {
			return more
		}
	}
	return true
}

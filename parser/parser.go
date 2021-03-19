package parser

import (
	"encoding/base64"
	"fmt"
	"github.com/zishang520/engine.io/bytes"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/types"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Callback func(*types.Packet, int, int) bool

/**
 * Current protocol version.
 */
const Protocol = 3

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
)

/**
 * Encodes a packet.
 *
 *     <packet type id> [ <data> ]
 *
 * Example:
 *
 *     5hello world
 *     3
 *     4
 *
 * Binary is encoded in an identical principle
 *
 * @api public
 */

func EncodePacket(packet *types.Packet, supportsBinary bool) (*bytes.Buffer, error) {
	encode := bytes.NewBuffer(nil)

	if packet == nil {
		return encode, errors.New(`Packet is nil`)
	}

	if c, ok := packet.Data.(io.Closer); ok {
		defer c.Close()
	}

	Type, _type_ok := PACKET_TYPES[packet.Type]
	if !_type_ok {
		return encode, errors.New(`Packet Type error`)
	}

	switch v := packet.Data.(type) {
	case *bytes.StringBuffer:
		encode.WriteByte(PACKET_TYPES[packet.Type])
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

func hasBinary(packets []*types.Packet) bool {
	if len(packets) == 0 {
		return false
	}
	for _, packet := range packets {
		switch packet.Data.(type) {
		case *strings.Reader:
		case io.WriterTo:
			return true
		case nil:
		default:
		}
	}

	return false
}

/**
 * Encodes multiple messages (payload).
 *
 *     <length>:data
 *
 * Example:
 *
 *     11:hello world2:hi
 *
 * If any contents are binary, they will be encoded as base64 strings. Base64
 * encoded strings are marked with a b before the length specifier
 *
 * @param {slice} packets
 * @api public
 */

func EncodePayload(packets []*types.Packet, supportsBinary bool) (*bytes.Buffer, error) {
	isBinary := hasBinary(packets)
	if supportsBinary && isBinary {
		return EncodePayloadAsBinary(packets)
	}

	enPayload := bytes.NewBuffer(nil)

	if len(packets) == 0 {
		enPayload.WriteString(`0:`)
		return enPayload, nil
	}

	for _, packet := range packets {
		if !isBinary {
			supportsBinary = false
		}
		if buf, err := EncodePacket(packet, supportsBinary, false); err != nil {
			return enPayload, err
		} else {
			enPayload.WriteString(fmt.Sprintf(`%d:%s`, Utf16Count(buf.Bytes()), buf.String()))
		}
	}

	return enPayload, nil
}

func encodeOneBinaryPacket(packet *types.Packet) (*bytes.Buffer, error) {
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
		buf.WriteTo(NewUtf8Encoder(binarypacket))
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

func EncodePayloadAsBinary(packets []*types.Packet) (*bytes.Buffer, error) {
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

func DecodePayload(data io.Reader, callback Callback) bool {
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

func DecodePayloadAsBinary(data io.Reader, callback Callback) bool {
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

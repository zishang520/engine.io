package parser

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/types"
	"io"
	"strings"
	"unicode/utf8"
)

/**
 * Current protocol version.
 */
const Protocol = 3

/**
 * Packet types.
 */
var (
	packets map[string]byte = map[string]byte{
		"open":    '0', // non-ws
		"close":   '1', // non-ws
		"ping":    '2',
		"pong":    '3',
		"message": '4',
		"upgrade": '5',
		"noop":    '6',
	}
	packetslist map[byte]string = map[byte]string{'0': "open", '1': "close", '2': "ping", '3': "pong", '4': "message", '5': "upgrade", '6': "noop"}

	EMPTY_BUFFER *bytes.Buffer = bytes.NewBuffer(nil)

	errPacket = &types.Packet{Type: `error`, Data: strings.NewReader(`parser error`)}
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

func EncodePacket(packet *types.Packet, supportsBinary bool, utf8encode bool) (*bytes.Buffer, error) {
	encode := bytes.NewBuffer(nil)

	if c, ok := packet.Data.(io.Closer); ok {
		defer c.Close()
	}

	switch v := packet.Data.(type) {
	case *strings.Reader:
		encode.WriteByte(packets[packet.Type])
		if utf8encode {
			v.WriteTo(NewUtf8Encoder(&Opts{Strict: false}, encode))
		} else {
			v.WriteTo(encode)
		}
	case io.WriterTo:
		if !supportsBinary {
			encode.Write([]byte{'b', packets[packet.Type]})
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			v.WriteTo(b64)
		} else {
			encode.WriteByte(packets[packet.Type] - '0')
			v.WriteTo(encode)
		}
	case nil:
		encode.WriteByte(packets[packet.Type])
	default:
		return encode, errors.New(`unknown packet.Data type`)
	}

	return encode, nil
}

/**
 * Decodes a packet. Data also available as an ArrayBuffer if requested.
 *
 * @return {Object} with `type` and `data` (if any)
 * @api public
 */

func DecodePacket(data io.Reader, utf8decode bool) (*types.Packet, error) {
	if data == nil {
		return errPacket, nil
	}
	if c, ok := data.(io.Closer); ok {
		defer c.Close()
	}

	msgType := []byte{0xFF}
	if _, err := data.Read(msgType); err != nil {
		return errPacket, err
	}

	decode := bytes.NewBuffer(nil)
	switch v := data.(type) {
	case *strings.Reader:
		if msgType[0] == 'b' {
			if _, err := v.Read(msgType); err != nil {
				return errPacket, err
			}
			packetType, ok := packetslist[msgType[0]]
			if !ok {
				return errPacket, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
			}
			decode.ReadFrom(base64.NewDecoder(base64.StdEncoding, v))
			return &types.Packet{
				Type: packetType,
				Data: decode,
			}, nil
		}
		packetType, ok := packetslist[msgType[0]]
		if !ok {
			return errPacket, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
		}
		if utf8decode {
			decode.ReadFrom(NewUtf8Decoder(&Opts{Strict: false}, v))
		} else {
			decode.ReadFrom(v)
		}
		return &types.Packet{
			Type: packetType,
			Data: decode,
		}, nil
	case io.Reader:
		packetType, ok := packetslist[msgType[0]]
		if !ok {
			return errPacket, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
		}
		decode.ReadFrom(v)
		return &types.Packet{
			Type: packetType,
			Data: decode,
		}, nil
	default:
	}

	return errPacket, errors.New(`parser error`)
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
 * @param {Array} packets
 * @api private
 */

func EncodePayload(packets []*types.Packet, supportsBinary bool) (*bytes.Buffer, error) {
	if supportsBinary && hasBinary(packets) {
		return EncodePayloadAsBinary(packets)
	}

	enPayload := bytes.NewBuffer(nil)

	if len(packets) == 0 {
		enPayload.WriteString(`0:`)
		return enPayload, nil
	}

	for _, packet := range packets {
		if buf, err := EncodePacket(packet, supportsBinary, false); err != nil {
			return enPayload, err
		} else {
			enPayload.WriteString(fmt.Sprintf(`%d:%s`, utf8.RuneCount(buf.Bytes()), buf.String()))
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
		encodingLength := fmt.Sprintf(`%d`, utf8.RuneCount(buf.Bytes()))
		binarypacket.WriteByte(0)
		for i := 0; i < len(encodingLength); i++ {
			binarypacket.WriteByte(encodingLength[i] - '0')
		}
		binarypacket.WriteByte(0xFF)
		for buf.Len() > 0 {
			r, _, e := buf.ReadRune()
			if e != nil && e != io.EOF {
				return binarypacket, e
			}
			binarypacket.WriteByte(byte(r))
		}
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
	if len(packets) == 0 {
		return EMPTY_BUFFER, nil
	}
	enPayload := bytes.NewBuffer(nil)

	for _, packet := range packets {
		if buf, err := encodeOneBinaryPacket(packet); err != nil {
			return enPayload, err
		} else {
			enPayload.ReadFrom(buf)
		}
	}

	return enPayload, nil
}

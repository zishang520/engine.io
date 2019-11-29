package parser

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/types"
	"io"
	"strings"
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
		"open":    0x30, // non-ws
		"close":   0x31, // non-ws
		"ping":    0x32,
		"pong":    0x33,
		"message": 0x34,
		"upgrade": 0x35,
		"noop":    0x36,
	}
	packetslist map[byte]string = map[byte]string{0x30: "open", 0x31: "close", 0x32: "ping", 0x33: "pong", 0x34: "message", 0x35: "upgrade", 0x36: "noop"}

	EMPTY_BUFFER *bytes.Buffer = new(bytes.Buffer)

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
		} else {
			encode.WriteByte(packets[packet.Type])
		}
		if !supportsBinary {
			b64 := base64.NewEncoder(base64.StdEncoding, encode)
			defer b64.Close()

			v.WriteTo(b64)
		} else {
			v.WriteTo(encode)
		}

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

	msgType := []byte{0XFF}
	if _, err := v.Read(msgType); err != nil {
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
	case io.Reader:
		packetType, ok := packetslist[msgType[0]]
		if !ok {
			return errPacket, errors.New(fmt.Sprintf(`Parsing error, unknown data type [%c]`, msgType[0]))
		}
		decode.ReadFrom(v)
	default:
		return errPacket, errors.New(`unknown data type`)
	}

	return &types.Packet{
		Type: packetType,
		Data: decode,
	}, nil
}

// func tryDecode(data ) {
//   try {
//     data = utf8.decode(data, { strict: false });
//   } catch (e) {
//     return false;
//   }
//   return data;
// }

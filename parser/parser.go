package parser

import (
	"bytes"
	"encoding/base64"
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

	err = types.Packet{Type: `error`, Data: strings.NewReader(`parser error`)}
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
			v.WriteTo(Utf8NewEncoder(&Opts{Strict: false}, encode))
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
	}

	return encode, nil
}

/**
 * Decodes a packet. Data also available as an ArrayBuffer if requested.
 *
 * @return {Object} with `type` and `data` (if any)
 * @api public
 */

// func DecodePacket(data io.Reader, binaryType string, utf8decode bool) (*types.Packet, error) {
// 	if data == nil {
// 		return err
// 	}
// 	if c, ok := data.(io.Closer); ok {
// 		defer c.Close()
// 	}
// 	msgType := []byte{0XFF}
// 	decode := bytes.NewBuffer(nil)

// 	switch v := data.(type) {
// 	case *bytes.Buffer:
// 		dataByte = v.Bytes()
// 	case *bytes.Reader:
// 		buf := new(bytes.Buffer)
// 		if _, err := buf.ReadFrom(v); err != nil {
// 			return nil, err
// 		}
// 		dataByte = buf.Bytes()
// 	case *strings.Reader:
// 		if _, err := v.Read(msgType); err != nil {
// 			return nil, err
// 		}
// 		if msgType[0] == 'b' {
// 			if _, err := v.Read(msgType); err != nil {
// 				return nil, err
// 			}
// 			Type := packetslist[msgType[0]]
// 			buf := new(bytes.Buffer)
// 			if _, err := buf.ReadFrom(v); err != nil {
// 				return nil, err
// 			}
// 			b64 = base64.NewDecoder(base64.StdEncoding, decode)
// 			defer b64.Close()
// 			b64.Write(buf.Bytes())
// 			return types.Packet{
// 				Type: Type,
// 				Data: decode,
// 			}
// 		}
// 		if utf8decode {
// 			data := tryDecode(data)
// 			if data == false {
// 				return err
// 			}
// 		}
// 	// buf := new(bytes.Buffer)
// 	// if _, err := buf.ReadFrom(v); err != nil {
// 	// 	return nil, err
// 	// }
// 	// dataByte = Utf8encodeByte(buf.Bytes(), &Opts{Strict: false})
// 	default:
// 	}
// 	//   var type;

// 	// String data
// 	//   if (typeof data === 'string') {

// 	//     type = data.charAt(0);

// 	//     if (type === 'b') {
// 	//       return exports.decodeBase64Packet(data.substr(1), binaryType);
// 	//     }

// 	//     if (utf8decode) {
// 	//       data = tryDecode(data);
// 	//       if (data === false) {
// 	//         return err;
// 	//       }
// 	//     }

// 	//     if (Number(type) != type || !packetslist[type]) {
// 	//       return err;
// 	//     }

// 	//     if (data.length > 1) {
// 	//       return { type: packetslist[type], data: data.substring(1) };
// 	//     } else {
// 	//       return { type: packetslist[type] };
// 	//     }
// 	//   }

// 	//   // Binary data
// 	//   if (binaryType === 'arraybuffer') {
// 	//     // wrap Buffer/ArrayBuffer data into an Uint8Array
// 	//     var intArray = new Uint8Array(data);
// 	//     type = intArray[0];
// 	//     return { type: packetslist[type], data: intArray.buffer.slice(1) };
// 	//   }

// 	//   if (data instanceof ArrayBuffer) {
// 	//     data = arrayBufferToBuffer(data);
// 	//   }
// 	//   type = data[0];
// 	//   return { type: packetslist[type], data: data.slice(1) };
// }

// func tryDecode(data ) {
//   try {
//     data = utf8.decode(data, { strict: false });
//   } catch (e) {
//     return false;
//   }
//   return data;
// }

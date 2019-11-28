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
 * @api private
 */

func EncodePacket(packet types.Packet, supportsBinary bool, utf8encode bool) (*bytes.Buffer, error) {
	encode := bytes.NewBuffer(nil)
	if !supportsBinary {
		encode.Write([]byte{'b', packets[packet.Type]})
	} else {
		encode.WriteByte(packets[packet.Type])
	}
	var dataByte []byte

	if c, ok := packet.Data.(io.Closer); ok {
		defer c.Close()
	}

	switch v := packet.Data.(type) {
	case *bytes.Buffer:
		dataByte = v.Bytes()
	case *bytes.Reader, *strings.Reader:
		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(v); err != nil {
			return nil, err
		}
		dataByte = buf.Bytes()
	default:
	}

	if !supportsBinary {
		b64 := base64.NewEncoder(base64.StdEncoding, encode)
		defer b64.Close()

		b64.Write(dataByte)
	} else {
		encode.Write(dataByte)
	}

	return encode, nil
}

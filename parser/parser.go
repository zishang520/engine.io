package parser

import (
	"bytes"
	"encoding/base64"
	"fmt"
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
	packets map[string]int = map[string]int{
		"open":    0, // non-ws
		"close":   1, // non-ws
		"ping":    2,
		"pong":    3,
		"message": 4,
		"upgrade": 5,
		"noop":    6,
	}
	packetslist map[int]string = map[int]string{0: "open", 1: "close", 2: "ping", 3: "pong", 4: "message", 5: "upgrade", 6: "noop"}

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
		encode.WriteString(fmt.Sprintf("b%d", packets[packet.Type]))
	} else {
		encode.WriteString(fmt.Sprintf("%d", packets[packet.Type]))
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

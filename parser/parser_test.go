package parser

import (
	"bytes"
	"testing"

	"github.com/zishang520/engine.io/packet"
)

func TestParserv3(t *testing.T) {
	p := Parserv3()

	t.Run("Protocol", func(t *testing.T) {
		if protocol := p.Protocol(); protocol != 3 {
			t.Fatalf(`*Parserv3.Protocol() = %d, want match for %d`, protocol, 3)
		}
	})

	t.Run("EncodePacketByte", func(t *testing.T) {
		data, err := p.EncodePacket(&packet.Packet{
			Type:    packet.OPEN,
			Data:    bytes.NewBuffer([]byte("ABC")),
			Options: nil,
		}, true, false)

		if err != nil {
			t.Fatal("Error with EncodePacket:", err)
		}

		if b := data.Bytes(); !bytes.Equal(b, []byte{0x00, 65, 66, 67}) {
			t.Fatalf(`EncodePacket value not as expected: %v, want match for %v`, b, []byte{0x00, 65, 66, 67})
		}

		data, err = p.EncodePacket(&packet.Packet{
			Type:    packet.OPEN,
			Data:    bytes.NewBuffer([]byte("ABC")),
			Options: nil,
		}, false, false)

		if err != nil {
			t.Fatal("Error with EncodePacket:", err)
		}

		if b := data.String(); b != "b0QUJD" {
			t.Fatalf(`EncodePacket value not as expected: %s, want match for %s`, b, "b0QUJD")
		}
	})
}

func TestParserv4(t *testing.T) {

}

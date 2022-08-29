package parser

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/types"
)

func TestParserv3(t *testing.T) {
	p := Parserv3()

	t.Run("Protocol", func(t *testing.T) {
		if protocol := p.Protocol(); protocol != 3 {
			t.Fatalf(`*Parserv3.Protocol() = %d, want match for %d`, protocol, 3)
		}
	})

	t.Run("EncodePacket/Byte", func(t *testing.T) {
		data, err := p.EncodePacket(&packet.Packet{
			Type:    packet.OPEN,
			Data:    bytes.NewBuffer([]byte("ABC")),
			Options: nil,
		}, true)

		if err != nil {
			t.Fatal("Error with EncodePacket:", err)
		}
		check := []byte{0x00, 65, 66, 67}
		if b := data.Bytes(); !bytes.Equal(b, check) {
			t.Fatalf(`EncodePacket value not as expected: %v, want match for %v`, b, check)
		}
	})

	t.Run("EncodePacket/Byte/Base64", func(t *testing.T) {
		data, err := p.EncodePacket(&packet.Packet{
			Type:    packet.OPEN,
			Data:    bytes.NewBuffer([]byte("ABC")),
			Options: nil,
		}, false)

		if err != nil {
			t.Fatal("Error with EncodePacket:", err)
		}
		check1 := "b0QUJD"
		if b := data.String(); b != check1 {
			t.Fatalf(`EncodePacket value not as expected: %s, want match for %s`, b, check1)
		}

	})

	t.Run("EncodePacket/String", func(t *testing.T) {
		data, err := p.EncodePacket(&packet.Packet{
			Type:    packet.OPEN,
			Data:    strings.NewReader("test测试中文和表情字符❤️🧡💛🧓🏾💟"),
			Options: nil,
		}, false, false)

		if err != nil {
			t.Fatal("Error with EncodePacket:", err)
		}
		check2 := "0test测试中文和表情字符❤️🧡💛🧓🏾💟"
		if b := data.String(); b != check2 {
			t.Fatalf(`EncodePacket value not as expected: %s, want match for %s`, b, check2)
		}
	})

	t.Run("EncodePacket/String/Utf8encode", func(t *testing.T) {
		data, err := p.EncodePacket(&packet.Packet{
			Type:    packet.OPEN,
			Data:    strings.NewReader("test测试中文和表情字符❤️🧡💛🧓🏾💟"),
			Options: nil,
		}, false, true)

		if err != nil {
			t.Fatal("Error with EncodePacket:", err)
		}
		check3 := []byte{PACKET_TYPES[packet.OPEN], 116, 101, 115, 116, 195, 166, 194, 181, 194, 139, 195, 168, 194, 175, 194, 149, 195, 164, 194, 184, 194, 173, 195, 166, 194, 150, 194, 135, 195, 165, 194, 146, 194, 140, 195, 168, 194, 161, 194, 168, 195, 166, 194, 131, 194, 133, 195, 165, 194, 173, 194, 151, 195, 167, 194, 172, 194, 166, 195, 162, 194, 157, 194, 164, 195, 175, 194, 184, 194, 143, 195, 176, 194, 159, 194, 167, 194, 161, 195, 176, 194, 159, 194, 146, 194, 155, 195, 176, 194, 159, 194, 167, 194, 147, 195, 176, 194, 159, 194, 143, 194, 190, 195, 176, 194, 159, 194, 146, 194, 159}
		if b := data.Bytes(); !bytes.Equal(b, check3) {
			t.Fatalf(`EncodePacket value not as expected: %v, want match for %v`, b, check3)
		}
	})

	t.Run("DecodePacket/Byte/Base64", func(t *testing.T) {
		pack, err := p.DecodePacket(types.NewStringBufferString("b1QUJD"))

		if err != nil {
			t.Fatal("Error with DecodePacket:", err)
		}

		if pack.Type != packet.CLOSE {
			t.Fatalf(`DecodePacket *Packet.Type value not as expected: %q, want match for %q`, pack.Type, packet.CLOSE)
		}

		if pack.Data == nil {
			t.Fatal(`DecodePacket *Packet.Data value must not be nil`)
		}

		if c, ok := pack.Data.(io.Closer); ok {
			defer c.Close()
		}

		buf, err := types.NewBytesBufferReader(pack.Data)
		if err != nil {
			t.Fatal("io.Reader data read failed:", err)
		}

		check := "ABC"

		if b := buf.String(); b != check {
			t.Fatalf(`DecodePacket *Packet.Data value not as expected: %q, want match for %q`, b, check)
		}
	})

	t.Run("DecodePacket/Byte", func(t *testing.T) {
		pack, err := p.DecodePacket(types.NewBytesBuffer([]byte{PACKET_TYPES[packet.CLOSE] - '0', 65, 66, 67}))

		if err != nil {
			t.Fatal("Error with DecodePacket:", err)
		}

		if pack.Type != packet.CLOSE {
			t.Fatalf(`DecodePacket *Packet.Type value not as expected: %q, want match for %q`, pack.Type, packet.CLOSE)
		}

		if pack.Data == nil {
			t.Fatal(`DecodePacket *Packet.Data value must not be nil`)
		}

		if c, ok := pack.Data.(io.Closer); ok {
			defer c.Close()
		}

		buf, err := types.NewBytesBufferReader(pack.Data)
		if err != nil {
			t.Fatal("io.Reader data read failed:", err)
		}

		check := []byte{65, 66, 67}

		if b := buf.Bytes(); !bytes.Equal(b, check) {
			t.Fatalf(`DecodePacket *Packet.Data value not as expected: %v, want match for %v`, b, check)
		}
	})

	t.Run("DecodePacket/String/Utf8decode", func(t *testing.T) {
		pack, err := p.DecodePacket(types.NewStringBuffer([]byte{PACKET_TYPES[packet.PING], 116, 101, 115, 116, 195, 166, 194, 181, 194, 139, 195, 168, 194, 175, 194, 149, 195, 164, 194, 184, 194, 173, 195, 166, 194, 150, 194, 135, 195, 165, 194, 146, 194, 140, 195, 168, 194, 161, 194, 168, 195, 166, 194, 131, 194, 133, 195, 165, 194, 173, 194, 151, 195, 167, 194, 172, 194, 166, 195, 162, 194, 157, 194, 164, 195, 175, 194, 184, 194, 143, 195, 176, 194, 159, 194, 167, 194, 161, 195, 176, 194, 159, 194, 146, 194, 155, 195, 176, 194, 159, 194, 167, 194, 147, 195, 176, 194, 159, 194, 143, 194, 190, 195, 176, 194, 159, 194, 146, 194, 159}), true)

		if err != nil {
			t.Fatal("Error with DecodePacket:", err)
		}

		if pack.Type != packet.PING {
			t.Fatalf(`DecodePacket *Packet.Type value not as expected: %q, want match for %q`, pack.Type, packet.PING)
		}

		if pack.Data == nil {
			t.Fatal(`DecodePacket *Packet.Data value must not be nil`)
		}

		if c, ok := pack.Data.(io.Closer); ok {
			defer c.Close()
		}

		buf, err := types.NewBytesBufferReader(pack.Data)
		if err != nil {
			t.Fatal("io.Reader data read failed:", err)
		}

		check := "test测试中文和表情字符❤️🧡💛🧓🏾💟"

		if b := buf.String(); b != check {
			t.Fatalf(`DecodePacket *Packet.Data value not as expected: %s, want match for %s`, b, check)
		}
	})

	t.Run("DecodePacket/String", func(t *testing.T) {
		pack, err := p.DecodePacket(types.NewStringBufferString("2test测试中文和表情字符❤️🧡💛🧓🏾💟"))

		if err != nil {
			t.Fatal("Error with DecodePacket:", err)
		}

		if pack.Type != packet.PING {
			t.Fatalf(`DecodePacket *Packet.Type value not as expected: %q, want match for %q`, pack.Type, packet.PING)
		}

		if pack.Data == nil {
			t.Fatal(`DecodePacket *Packet.Data value must not be nil`)
		}

		if c, ok := pack.Data.(io.Closer); ok {
			defer c.Close()
		}

		buf, err := types.NewBytesBufferReader(pack.Data)
		if err != nil {
			t.Fatal("io.Reader data read failed:", err)
		}

		check := "test测试中文和表情字符❤️🧡💛🧓🏾💟"

		if b := buf.String(); b != check {
			t.Fatalf(`DecodePacket *Packet.Data value not as expected: %s, want match for %s`, b, check)
		}
	})

	t.Run("EncodePayload/Base64", func(t *testing.T) {
		data, err := p.EncodePayload(
			[]*packet.Packet{
				&packet.Packet{
					Type:    packet.OPEN,
					Data:    bytes.NewBuffer([]byte("ABC")),
					Options: nil,
				},
				&packet.Packet{
					Type:    packet.CLOSE,
					Data:    strings.NewReader("test测试中文和表情字符❤️🧡💛🧓🏾💟"),
					Options: nil,
				},
			}, false)

		if err != nil {
			t.Fatal("Error with EncodePayload:", err)
		}
		check1 := "6:b0QUJD26:1test测试中文和表情字符❤️🧡💛🧓🏾💟"
		if b := data.String(); b != check1 {
			t.Fatalf(`EncodePayload value not as expected: %s, want match for %s`, b, check1)
		}
	})

	t.Run("EncodePayload", func(t *testing.T) {
		data, err := p.EncodePayload(
			[]*packet.Packet{
				&packet.Packet{
					Type:    packet.OPEN,
					Data:    bytes.NewBuffer([]byte("ABC")),
					Options: nil,
				},
				&packet.Packet{
					Type:    packet.CLOSE,
					Data:    strings.NewReader("test测试中文和表情字符❤️🧡💛🧓🏾💟"),
					Options: nil,
				},
			}, true)

		if err != nil {
			t.Fatal("Error with EncodePayload:", err)
		}
		check := []byte{1, 4, 255, 0, 65, 66, 67, 0, 5, 8, 255, 49, 116, 101, 115, 116, 195, 131, 194, 166, 195, 130, 194, 181, 195, 130, 194, 139, 195, 131, 194, 168, 195, 130, 194, 175, 195, 130, 194, 149, 195, 131, 194, 164, 195, 130, 194, 184, 195, 130, 194, 173, 195, 131, 194, 166, 195, 130, 194, 150, 195, 130, 194, 135, 195, 131, 194, 165, 195, 130, 194, 146, 195, 130, 194, 140, 195, 131, 194, 168, 195, 130, 194, 161, 195, 130, 194, 168, 195, 131, 194, 166, 195, 130, 194, 131, 195, 130, 194, 133, 195, 131, 194, 165, 195, 130, 194, 173, 195, 130, 194, 151, 195, 131, 194, 167, 195, 130, 194, 172, 195, 130, 194, 166, 195, 131, 194, 162, 195, 130, 194, 157, 195, 130, 194, 164, 195, 131, 194, 175, 195, 130, 194, 184, 195, 130, 194, 143, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 167, 195, 130, 194, 161, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 146, 195, 130, 194, 155, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 167, 195, 130, 194, 147, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 143, 195, 130, 194, 190, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 146, 195, 130, 194, 159}

		if b := data.Bytes(); !bytes.Equal(b, check) {
			t.Fatalf(`DecodePacket *Packet.Data value not as expected: %v, want match for %v`, b, check)
		}
	})

	t.Run("DecodePayload/Base64", func(t *testing.T) {
		packs := p.DecodePayload(types.NewStringBufferString("6:b0QUJD26:1test测试中文和表情字符❤️🧡💛🧓🏾💟"))

		if l := len(packs); l != 2 {
			t.Fatalf(`*len(packs) = %d, want match for %d`, l, 2)
		}

		func() {

			if tp := packs[0].Type; tp != packet.OPEN {
				t.Fatalf(`DecodePayload packs[0].Type value not as expected: %q, want match for %q`, tp, packet.OPEN)
			}

			if packs[0].Data == nil {
				t.Fatal(`DecodePacket packs[0]..Data value must not be nil`)
			}

			if c, ok := packs[0].Data.(io.Closer); ok {
				defer c.Close()
			}

			buf, err := types.NewBytesBufferReader(packs[0].Data)
			if err != nil {
				t.Fatal("packs[0] io.Reader data read failed:", err)
			}

			check := []byte{65, 66, 67}

			if b := buf.Bytes(); !bytes.Equal(b, check) {
				t.Fatalf(`DecodePacket packs[0]..Data value not as expected: %v, want match for %v`, b, check)
			}
		}()

		func() {

			if tp := packs[1].Type; tp != packet.CLOSE {
				t.Fatalf(`DecodePayload packs[1].Type value not as expected: %q, want match for %q`, tp, packet.CLOSE)
			}

			if packs[1].Data == nil {
				t.Fatal(`DecodePacket packs[1].Data value must not be nil`)
			}

			if c, ok := packs[1].Data.(io.Closer); ok {
				defer c.Close()
			}

			buf, err := types.NewBytesBufferReader(packs[1].Data)
			if err != nil {
				t.Fatal("io.Reader data read failed:", err)
			}

			check := "test测试中文和表情字符❤️🧡💛🧓🏾💟"

			if b := buf.String(); b != check {
				t.Fatalf(`DecodePacket packs[1].Data value not as expected: %s, want match for %s`, b, check)
			}
		}()
	})

	t.Run("DecodePayload", func(t *testing.T) {
		packs := p.DecodePayload(types.NewBytesBuffer([]byte{1, 4, 255, 0, 65, 66, 67, 0, 5, 8, 255, 49, 116, 101, 115, 116, 195, 131, 194, 166, 195, 130, 194, 181, 195, 130, 194, 139, 195, 131, 194, 168, 195, 130, 194, 175, 195, 130, 194, 149, 195, 131, 194, 164, 195, 130, 194, 184, 195, 130, 194, 173, 195, 131, 194, 166, 195, 130, 194, 150, 195, 130, 194, 135, 195, 131, 194, 165, 195, 130, 194, 146, 195, 130, 194, 140, 195, 131, 194, 168, 195, 130, 194, 161, 195, 130, 194, 168, 195, 131, 194, 166, 195, 130, 194, 131, 195, 130, 194, 133, 195, 131, 194, 165, 195, 130, 194, 173, 195, 130, 194, 151, 195, 131, 194, 167, 195, 130, 194, 172, 195, 130, 194, 166, 195, 131, 194, 162, 195, 130, 194, 157, 195, 130, 194, 164, 195, 131, 194, 175, 195, 130, 194, 184, 195, 130, 194, 143, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 167, 195, 130, 194, 161, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 146, 195, 130, 194, 155, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 167, 195, 130, 194, 147, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 143, 195, 130, 194, 190, 195, 131, 194, 176, 195, 130, 194, 159, 195, 130, 194, 146, 195, 130, 194, 159}))

		if l := len(packs); l != 2 {
			t.Fatalf(`*len(packs) = %d, want match for %d`, l, 2)
		}

		func() {

			if tp := packs[0].Type; tp != packet.OPEN {
				t.Fatalf(`DecodePayload packs[0].Type value not as expected: %q, want match for %q`, tp, packet.OPEN)
			}

			if packs[0].Data == nil {
				t.Fatal(`DecodePacket packs[0]..Data value must not be nil`)
			}

			if c, ok := packs[0].Data.(io.Closer); ok {
				defer c.Close()
			}

			buf, err := types.NewBytesBufferReader(packs[0].Data)
			if err != nil {
				t.Fatal("packs[0] io.Reader data read failed:", err)
			}

			check := []byte{65, 66, 67}

			if b := buf.Bytes(); !bytes.Equal(b, check) {
				t.Fatalf(`DecodePacket packs[0]..Data value not as expected: %v, want match for %v`, b, check)
			}
		}()

		func() {

			if tp := packs[1].Type; tp != packet.CLOSE {
				t.Fatalf(`DecodePayload packs[1].Type value not as expected: %q, want match for %q`, tp, packet.CLOSE)
			}

			if packs[1].Data == nil {
				t.Fatal(`DecodePacket packs[1].Data value must not be nil`)
			}

			if c, ok := packs[1].Data.(io.Closer); ok {
				defer c.Close()
			}

			buf, err := types.NewBytesBufferReader(packs[1].Data)
			if err != nil {
				t.Fatal("io.Reader data read failed:", err)
			}

			check := "test测试中文和表情字符❤️🧡💛🧓🏾💟"

			if b := buf.String(); b != check {
				t.Fatalf(`DecodePacket packs[1].Data value not as expected: %s, want match for %s`, b, check)
			}
		}()
	})

}

func TestParserv4(t *testing.T) {

}

package main

import (
	"bytes"
	"fmt"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"strings"
)

func main() {
	buf, err := parser.EncodePayload([]*types.Packet{
		&types.Packet{
			Type: "ping",
			Data: strings.NewReader(`[]byte{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`),
		},
		&types.Packet{
			Type: "close",
			Data: strings.NewReader(`[]byte{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`),
		},
		&types.Packet{
			Type: "noop",
			Data: strings.NewReader(`[]byte{0, 1你好呀, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`),
		},
	}, false)
	bufs, err := parser.EncodePacket(&types.Packet{
		Type: "ping",
		Data: bytes.NewBuffer([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}),
	}, true, false)
	bufss, err := parser.EncodePacket(&types.Packet{
		Type: "close",
	}, true, true)
	bufsss, err := parser.EncodePacket(&types.Packet{
		Type: "pong",
		Data: strings.NewReader(string([]rune{0x31, 0x31, 0xF800, 2, 3, 4, 5, 6, 7, 8, 9})),
	}, true, true)
	dbufsss, err := parser.DecodePacket(strings.NewReader(string([]rune{0x31, 0x31, 0xF800, 2, 3, 4, 5, 6, 7, 8, 9})), false)
	dbufsssu, err := parser.DecodePacket(strings.NewReader(`b3`), true)
	fmt.Println(err)
	fmt.Println(buf)
	fmt.Println(bufs)
	fmt.Println(bufss)
	fmt.Println(bufsss)
	fmt.Println(dbufsss)
	fmt.Println(dbufsssu)
	fmt.Println(dbufsssu.Data)
	fmt.Println(dbufsss.Data)
}

package main

import (
	// "types"
	"fmt"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	// "strings"
)

func main() {
	buf, err := parser.EncodePayload([]*packet.Packet{
		&packet.Packet{
			Type: "ping",
			Data: types.NewBytesBuffer([]byte(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁wn5iB`)),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewBytesBuffer([]byte(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁wn5iB😀😁`)),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewBytesBuffer([]byte(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`)),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewBytesBuffer([]byte(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`)),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewStringBufferString(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewStringBufferString(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewStringBufferString(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewStringBufferString(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewStringBufferString(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
		},
		&packet.Packet{
			Type: "pong",
			Data: types.NewStringBufferString(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁9999999999999999999999999999😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
		},
		&packet.Packet{
			Type: "close",
			Data: types.NewStringBufferString(`😀😁😀😁你好呀, 2, 3, 4, 5,😀😀😁😀😁😀😀😁😀😁😀😀😁😀😁😀`),
		},
		&packet.Packet{
			Type: "ping",
			Data: types.NewStringBufferString(`[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`),
		},
		&packet.Packet{
			Type: "noop",
			Data: types.NewStringBufferString(`[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9wn5iB`),
		},
	})
	fmt.Println(err)
	fmt.Println(buf.String())
	x := parser.DecodePayload(buf)
	fmt.Println(x)
	bufs, errs := parser.EncodePayload(x)
	fmt.Println(errs)
	fmt.Println(bufs.String())

	// fmt.Println(parser.DecodePayload(strings.NewReader(`209:2😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁103:2[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl91:6`), func(a *packet.Packet, b int, c int) bool {
	// 	fmt.Println(a)
	// 	fmt.Println(b)
	// 	fmt.Println(c)
	// 	return true
	// }))
}

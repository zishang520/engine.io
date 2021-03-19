package main

import (
	// "bytes"
	"fmt"
	"github.com/zishang520/engine.io/bytes"
	// "strings"
)

func main() {
	// buf, err := parser.EncodePayload([]*types.Packet{
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "pong",
	// 		Data: strings.NewReader(`😀😁😀😁😀你好呀, 2, 3, 4, 5,😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁`),
	// 	},
	// 	&types.Packet{
	// 		Type: "close",
	// 		Data: bytes.NewReader([]byte(`😀😁😀😁你好呀, 2, 3, 4, 5,😀😀😁😀😁😀😀😁😀😁😀😀😁😀😁😀`)),
	// 	},
	// 	&types.Packet{
	// 		Type: "ping",
	// 		Data: strings.NewReader(`[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`),
	// 	},
	// 	&types.Packet{
	// 		Type: "noop",
	// 		Data: strings.NewReader(`你好呀, 2, 3, 4, 5,`),
	// 	},
	// }, true)
	// fmt.Println(err)
	// // fmt.Println(buf.String())
	// fmt.Println(parser.DecodePayload(buf, func(a *types.Packet, b int, c int) bool {
	// 	fmt.Println(a)
	// 	fmt.Println(b)
	// 	fmt.Println(c)
	// 	return true
	// }))
	// fmt.Println(parser.DecodePayload(strings.NewReader(buf.String()), func(a *types.Packet, b int, c int) bool {
	// 	fmt.Println(a)
	// 	fmt.Println(b)
	// 	fmt.Println(c)
	// 	return true
	// }))

	x := bytes.NewStringBuffer([]byte(`😀😁😀😁你好呀, 2, 3, 4, 5,😀😀😁😀😁😀😀😁😀😁😀😀😁😀😁😀`))
	x1 := bytes.NewBuffer([]byte(`😀😁😀😁你好呀, 2, 3, 4, 5,😀😀😁😀😁😀😀😁😀😁😀😀😁😀😁😀`))
	// fmt.Println(parser.DecodePayload(strings.NewReader(`209:2😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁103:2[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl91:6`), func(a *types.Packet, b int, c int) bool {
	// 	fmt.Println(a)
	// 	fmt.Println(b)
	// 	fmt.Println(c)
	// 	return true
	// }))
}

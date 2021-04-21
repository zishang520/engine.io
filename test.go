package main

import (
	"fmt"
	"github.com/zishang520/engine.io/types"
	"unicode/utf8"
)

func main() {

	b := []byte{
		72, 101, 108, 108, 111, 44, 32, 228, 184, 150, 231, 149, 140, 0x01,
	}

	for len(b) > 0 {
		r, size := utf8.DecodeLastRune(b)
		fmt.Printf("%c %v\n", r, size)

		b = b[:len(b)-size]
	}
	x := types.NewBytesBufferString("测")
	// fmt.Println(parser.DecodePayload(strings.NewReader(`209:2😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁103:2[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl91:6`), func(a *packet.Packet, b int, c int) bool {
	x.Write([]byte{1, 2})
	fmt.Println(x.ReadRune())
	fmt.Println(x.ReadRune())
	fmt.Println(x.ReadRune())
	// 	fmt.Println(b)
	// 	fmt.Println(c)
	// 	return true
	// }))
}

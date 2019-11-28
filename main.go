package main

import (
	"bytes"
	"fmt"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"strings"
	"unicode/utf8"
)

func main() {
	buf, err := parser.EncodePacket(&types.Packet{
		Type: "ping",
		Data: bytes.NewBuffer([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}),
	}, false, false)
	bufs, err := parser.EncodePacket(&types.Packet{
		Type: "ping",
		Data: bytes.NewBuffer([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}),
	}, true, false)
	bufss, err := parser.EncodePacket(&types.Packet{
		Type: "ping",
		Data: strings.NewReader(`[]byte{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}`),
	}, true, false)
	bufsss, err := parser.EncodePacket(&types.Packet{
		Type: "ping",
		Data: strings.NewReader(`[]byte{0, 1, 2, 3, 你好呀4, 5, 6, 7, 8, 9}`),
	}, true, true)
	fmt.Println(buf)
	fmt.Println(bufs)

	fmt.Println(bufss)
	fmt.Println(bufsss)
	fmt.Println(err)
	fmt.Println(byte(0) + '0')
	fmt.Println('b')
	fmt.Println(0x31)
	xxxx := []byte{0xFF}
	b, l := utf8.DecodeRuneInString(`b[]byte{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}`)
	fmt.Println(byte(b) == 'b')
	fmt.Println(xxxx)
	fmt.Println(l)
}

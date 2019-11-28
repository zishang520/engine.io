package main

import (
	"bytes"
	"fmt"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"strings"
)

func main() {
	buf, err := parser.EncodePacket(types.Packet{
		Type: "ping",
		Data: bytes.NewBuffer([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}),
	}, false, false)
	bufs, err := parser.EncodePacket(types.Packet{
		Type: "ping",
		Data: bytes.NewBuffer([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}),
	}, true, false)
	bufss, err := parser.EncodePacket(types.Packet{
		Type: "ping",
		Data: strings.NewReader(`[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}`),
	}, true, true)
	bufsss, err := parser.EncodePacket(types.Packet{
		Type: "ping",
		Data: strings.NewReader(`[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}`),
	}, false, true)
	fmt.Println(buf)
	fmt.Println(bufs)
	fmt.Println(bufss)
	fmt.Println(bufsss)
	fmt.Println(err)
	fmt.Println('3' + '0')
	fmt.Println(parser.EMPTY_BUFFER)
}

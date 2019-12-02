package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"strings"
)

func main() {
	buf, err := parser.EncodePayload([]*types.Packet{
		&types.Packet{
			Type: "ping",
			Data: bytes.NewBuffer([]byte(`[]byte{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`)),
		},
		&types.Packet{
			Type: "ping",
			Data: strings.NewReader(`[]byte{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`),
		},
		&types.Packet{
			Type: "noop",
			Data: strings.NewReader(`[]byte{0, 1你好呀, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`),
		},
	}, true)
	fmt.Println(err)
	fmt.Println(hex.EncodeToString(buf.Bytes()))
	fmt.Println(buf.String())
}

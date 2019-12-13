package main

import (
	"bytes"
	"fmt"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

func main() {
	buf, err := parser.EncodePayload([]*types.Packet{
		&types.Packet{
			Type: "ping",
			Data: strings.NewReader(`,你好呀, 2, 3, 4, 5, 6`),
		},
		&types.Packet{
			Type: "close",
			Data: bytes.NewReader([]byte(`xxxxxxx`)),
		},
		&types.Packet{
			Type: "noop",
			Data: strings.NewReader(`, 你好呀, 2, 3, 4,`),
		},
		&types.Packet{
			Type: "noop",
			Data: strings.NewReader(`, 你好呀, 2, 3, 4,`),
		},
	}, true)
	fmt.Println(buf)
	fmt.Println(err)
	fmt.Println(utf8.Valid(buf.Bytes()))
	fmt.Println(utf8.Valid(buf.Bytes()))
	fmt.Println(parser.DecodePayload(buf, func(a *types.Packet, b int, c int) bool {
		fmt.Println(a)
		fmt.Println(b)
		fmt.Println(c)
		return true
	}))

	boolss := parser.DecodePayload(strings.NewReader(`102:2[]byte{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9103:2[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl987:6[]by:te{0, 1你好呀, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl9`), func(a *types.Packet, b int, c int) bool {
		fmt.Println(a)
		fmt.Println(b)
		fmt.Println(c)
		return true
	})
	fmt.Println(boolss)

	SignalC := make(chan os.Signal)

	signal.Notify(SignalC, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range SignalC {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				os.Exit(0)
			}
		}
	}()
	for {
		time.Sleep(time.Second)
	}
}

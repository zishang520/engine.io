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

	fmt.Println(parser.DecodePayload(strings.NewReader(`35:2102:2[]byte{0, 1你好呀, 2, 3, 4, 5, 614:b1eHh4eHh4eA==1:61:6`), func(a *types.Packet, b int, c int) bool {
		fmt.Println(a)
		fmt.Println(b)
		fmt.Println(c)
		return true
	}))

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

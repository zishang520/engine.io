package main

import (
	// "types"
	"fmt"
	"github.com/zishang520/engine.io/packet"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	"github.com/zishang520/engine.io/utils"
	"regexp"
	"time"
)

var closeTimeoutTimer *utils.Timer = nil

var ti int = 0

func test() {
	if closeTimeoutTimer != nil {
		utils.ClearTimeOut(closeTimeoutTimer)
	}
	closeTimeoutTimer = utils.SetInterval(func() {
		utils.Log.Debug("2s后执行的代码")
	}, 2000*time.Millisecond)
}

func main() {
	buf, err := parser.ParserV4.EncodePayload([]*packet.Packet{
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
	x := parser.ParserV4.DecodePayload(buf)
	fmt.Println(x)
	bufs, errs := parser.ParserV4.EncodePayload(x)
	fmt.Println(errs)
	fmt.Println(parser.ParserV4.Protocol())
	fmt.Println(bufs.String())
	fmt.Println(utils.Base64Id.GenerateId(nil))
	fmt.Println(utils.Base64Id.GenerateId(nil))
	fmt.Println(utils.Base64Id.GenerateId(nil))
	utils.Log.DEBUG = true
	utils.Log.Debug("121212")
	r := regexp.MustCompile(`\\\\n`)
	utils.Log.Debug(r.ReplaceAllString("\n\\n\\\\n", `\n`))
	test()
	test()
	test()
	test()
	t := time.Duration(0)
	ops := types.InitConfig
	utils.Log.Debug("%v", ops.Cookie)
	utils.Log.Debug("%v", ops.PerMessageDeflate)
	utils.Log.Debug("%v", ops.UpgradeTimeout)
	ops.Assign(&types.Config{PingTimeout: &t, Cors: &types.Cors{}})
	utils.Log.Debug("%v", ops)
	time.Sleep(10 * time.Second)
	utils.ClearTimeOut(closeTimeoutTimer)
	time.Sleep(30 * time.Second)

	// fmt.Println(parser.DecodePayload(strings.NewReader(`209:2😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁😀😁103:2[]by:te{0, 1你好呀, 2, 3, 4, 5, 6, 7, 8, 9}b2W11ieXRlezAsIDHkvaDlpb3lkYAsIDIsIDMsIDQsIDUsIDYsIDcsIDgsIDl91:6`), func(a *packet.Packet, b int, c int) bool {
	// 	fmt.Println(a)
	// 	fmt.Println(b)
	// 	fmt.Println(c)
	// 	return true
	// }))
}

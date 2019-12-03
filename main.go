package main

import (
	// "bytes"
	"encoding/hex"
	"fmt"
	"github.com/zishang520/engine.io/parser"
	"github.com/zishang520/engine.io/types"
	// "strings"
)

func main() {
	buf, err := parser.EncodePayload([]*types.Packet{}, true)
	fmt.Println(err)
	fmt.Println(hex.EncodeToString(buf.Bytes()))
	fmt.Println(buf.String())
	fmt.Println([]rune(``))
	fmt.Println(string([]rune{0x3a}))
}

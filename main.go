package main

import (
	"bytes"
	"fmt"
)

type Opts struct {
	Strict bool
}

func checkScalarValue(codePoint rune, strict bool) bool {
	if codePoint >= 0xD800 && codePoint <= 0xDFFF {
		if strict {
			panic(fmt.Sprintf(`Lone surrogate U+%s is not a scalar value`, fmt.Sprintf("%X", codePoint)))
		}
		return false
	}
	return true
}

func utf8encode(str string, opts *Opts) string {
	if opts == nil {
		opts = &Opts{false}
	}
	strs := []byte(str)
	var buf bytes.Buffer
	for _, b := range strs {
		rb := rune(b)
		if !checkScalarValue(rb, opts.Strict) {
			rb = 0xFFFD
		}
		buf.WriteRune(rb)
	}
	return buf.String()
}

// var byteArray;
// var byteCount;
// var byteIndex;
func utf8decode(byteString string, opts *Opts) string {
	if opts == nil {
		opts = &Opts{false}
	}
	str := []rune(byteString)
	var buf bytes.Buffer
	for _, r := range str {
		if !checkScalarValue(r, opts.Strict) {
			r = 0xFFFD
		}
		buf.WriteByte(byte(r))
	}
	return buf.String()
}

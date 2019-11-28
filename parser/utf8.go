package parser

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

func Utf8encode(str string, opts *Opts) string {
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

func Utf8encodeByte(strs []byte, opts *Opts) []byte {
	if opts == nil {
		opts = &Opts{false}
	}
	var buf bytes.Buffer
	for _, b := range strs {
		rb := rune(b)
		if !checkScalarValue(rb, opts.Strict) {
			rb = 0xFFFD
		}
		buf.WriteRune(rb)
	}
	return buf.Bytes()
}

func Utf8decode(byteString string, opts *Opts) string {
	if opts == nil {
		opts = &Opts{false}
	}
	strs := []rune(byteString)
	var buf bytes.Buffer
	for _, r := range strs {
		if !checkScalarValue(r, opts.Strict) {
			r = 0xFFFD
		}
		buf.WriteByte(byte(r))
	}
	return buf.String()
}

package parser

import (
	"bytes"
	"fmt"
	"io"
)

type Opts struct {
	Strict bool
}

func Utf8encodeString(str string, opts *Opts) string {
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

func Utf8encodeByte(dst, src []byte, opts *Opts) int {
	if opts == nil {
		opts = &Opts{false}
	}
	var buf bytes.Buffer
	for _, b := range src {
		rb := rune(b)
		if !checkScalarValue(rb, opts.Strict) {
			rb = 0xFFFD
		}
		buf.WriteRune(rb)
	}
	l, err := buf.Read(dst)
	if err != nil {
		return 0
	}
	return l
}

func Utf8decodeString(byteString string, opts *Opts) string {
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

// bufferSize is the number of hexadecimal characters to buffer in encoder and decoder.
const bufferSize = 1024

type utf8encoder struct {
	opts *Opts
	w    io.Writer
	err  error
	out  [bufferSize]byte // output buffer
}

// NewEncoder returns an io.Writer that writes lowercase hexadecimal characters to w.
func Utf8NewEncoder(opts *Opts, w io.Writer) io.Writer {
	return &utf8encoder{opts: opts, w: w}
}

func (e *utf8encoder) Write(p []byte) (n int, err error) {
	for len(p) > 0 && e.err == nil {
		chunkSize := bufferSize / 2
		if len(p) < chunkSize {
			chunkSize = len(p)
		}

		encoded := Utf8encodeByte(e.out[:], p[:chunkSize], e.opts)
		_, e.err = e.w.Write(e.out[:encoded])
		n += chunkSize
		p = p[chunkSize:]
	}
	return n, e.err
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

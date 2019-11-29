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

func Utf8encodeBytes(dst, src []byte, opts *Opts) int {
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

func Utf8decodeBytes(dst, src []byte, opts *Opts) (int, error) {
	if opts == nil {
		opts = &Opts{false}
	}
	buf := bytes.NewReader(src)
	l := buf.Len()
	for i := 0; i < l; i++ {
		r, _, e := buf.ReadRune()
		if e != nil {
			return 0, e
		}
		if !checkScalarValue(r, opts.Strict) {
			r = 0xFFFD
		}
		dst[i] = byte(r)
	}
	return l, nil
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

// bufferSize is the number of hexadecimal characters to buffer in encoder and decoder.
const bufferSize = 1024

type utf8encoder struct {
	opts *Opts
	w    io.Writer
	err  error
	out  [bufferSize]byte // output buffer
}

// NewEncoder returns an io.Writer that writes lowercase hexadecimal characters to w.
func NewUtf8Encoder(opts *Opts, w io.Writer) io.Writer {
	return &utf8encoder{opts: opts, w: w}
}

func (e *utf8encoder) Write(p []byte) (n int, err error) {
	for len(p) > 0 && e.err == nil {
		chunkSize := bufferSize / 2
		if len(p) < chunkSize {
			chunkSize = len(p)
		}

		encoded := Utf8encodeBytes(e.out[:], p[:chunkSize], e.opts)
		_, e.err = e.w.Write(e.out[:encoded])
		n += chunkSize
		p = p[chunkSize:]
	}
	return n, e.err
}

type utf8decoder struct {
	opts *Opts
	r    io.Reader
	err  error
	in   []byte           // input buffer (encoded form)
	arr  [bufferSize]byte // backing array for in
}

// NewDecoder returns an io.Reader that decodes hexadecimal characters from r.
// NewDecoder expects that r contain only an even number of hexadecimal characters.
func NewUtf8Decoder(opts *Opts, r io.Reader) io.Reader {
	return &utf8decoder{opts: opts, r: r}
}

func (d *utf8decoder) Read(p []byte) (n int, err error) {
	// Fill internal buffer with sufficient bytes to decode
	if len(d.in) == 0 && d.err == nil {
		var numCopy, numRead int
		numCopy = copy(d.arr[:], d.in) // Copies either 0 or 1 bytes
		numRead, d.err = d.r.Read(d.arr[numCopy:])
		d.in = d.arr[:numCopy+numRead]
	}

	// Decode internal buffer into output buffer
	if numAvail := len(d.in); len(p) > numAvail {
		p = p[:numAvail]
	}
	numDec, err := Utf8decodeBytes(p, d.in[:len(p)], d.opts)
	d.in = d.in[numDec:]
	if err != nil {
		d.in, d.err = nil, err // Decode error; discard input remainder
	}

	if len(d.in) == 0 {
		return numDec, d.err // Only expose errors when buffer fully consumed
	}
	return numDec, nil
}

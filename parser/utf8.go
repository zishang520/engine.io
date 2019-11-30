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

func Utf8decodeBytes(dst, src []byte, opts *Opts) (ndst, nsrc int, err error) {
	if opts == nil {
		opts = &Opts{false}
	}
	buf := bytes.NewReader(src)
	for buf.Len() > 0 {
		r, l, e := buf.ReadRune()
		if e != nil && e != io.EOF {
			return ndst, nsrc, e
		}
		if !checkScalarValue(r, opts.Strict) {
			r = 0xFFFD
		}
		dst[ndst] = byte(r)
		nsrc += l
		ndst++
	}
	return
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

func NewUtf8Decoder(opts *Opts, r io.Reader) io.Reader {
	return &utf8decoder{opts: opts, r: r}
}

type utf8decoder struct {
	opts    *Opts
	err     error
	readErr error
	r       io.Reader
	buf     [bufferSize]byte // leftover input
	nbuf    int
	out     []byte // leftover decoded output
	outbuf  [bufferSize]byte
}

func (d *utf8decoder) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if d.err != nil {
		return 0, d.err
	}

	for {
		// Copy leftover output from last decode.
		if len(d.out) > 0 {
			n = copy(p, d.out)
			d.out = d.out[n:]
			return
		}

		// Decode leftover input from last read.
		var nn, nsrc, ndst int
		if d.nbuf > 0 {
			ndst, nsrc, d.err = Utf8decodeBytes(d.outbuf[0:], d.buf[0:d.nbuf], d.opts)
			if ndst > 0 {
				d.out = d.outbuf[0:ndst]
				d.nbuf = copy(d.buf[0:], d.buf[nsrc:d.nbuf])
				continue // copy out and return
			}
		}

		// Out of input, out of decoded output. Check errors.
		if d.err != nil {
			return 0, d.err
		}
		if d.readErr != nil {
			d.err = d.readErr
			return 0, d.err
		}

		// Read more data.
		nn, d.readErr = d.r.Read(d.buf[d.nbuf:])
		d.nbuf += nn
	}
}

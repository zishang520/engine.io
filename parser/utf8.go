package parser

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"
)

func Utf8encodeString(str string) string {
	buf := bytes.NewBuffer(nil)
	for _, b := range str {
		rb := rune(b)
		if !utf8.ValidRune(rb) {
			rb = 0xFFFD
		}
		buf.WriteRune(rb)
	}
	return buf.String()
}

func Utf8encodeBytes(dst, src []byte) int {
	buf := bytes.NewBuffer(nil)
	for _, b := range src {
		rb := rune(b)
		if !utf8.ValidRune(rb) {
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

func Utf8decodeString(byteString string) string {
	strs := strings.NewReader(byteString)
	buf := bytes.NewBuffer(nil)
	for strs.Len() > 0 {
		r, _, e := buf.ReadRune()
		if e == nil {
			if !utf8.ValidRune(r) {
				r = 0xFFFD
			}
			buf.WriteByte(byte(r))
		}
	}
	return buf.String()
}

func Utf8decodeBytes(dst, src []byte) (ndst, nsrc int, err error) {
	buf := bytes.NewReader(src)
	for buf.Len() > 0 {
		r, l, e := buf.ReadRune()
		if e != nil && e != io.EOF {
			return ndst, nsrc, e
		}
		if !utf8.ValidRune(r) {
			r = 0xFFFD
		}
		dst[ndst] = byte(r)
		nsrc += l
		ndst++
	}
	return
}

// bufferSize is the number of hexadecimal characters to buffer in encoder and decoder.
const bufferSize = 1024

type utf8encoder struct {
	w   io.Writer
	err error
	out [bufferSize]byte // output buffer
}

// NewEncoder returns an io.Writer that writes lowercase hexadecimal characters to w.
func NewUtf8Encoder(w io.Writer) io.Writer {
	return &utf8encoder{w: w}
}

func (e *utf8encoder) Write(p []byte) (n int, err error) {
	for len(p) > 0 && e.err == nil {
		chunkSize := bufferSize / 2
		if len(p) < chunkSize {
			chunkSize = len(p)
		}

		encoded := Utf8encodeBytes(e.out[:], p[:chunkSize])
		_, e.err = e.w.Write(e.out[:encoded])
		n += chunkSize
		p = p[chunkSize:]
	}
	return n, e.err
}

type utf8decoder struct {
	err     error
	readErr error
	r       io.Reader
	buf     [bufferSize]byte // leftover input
	nbuf    int
	out     []byte // leftover decoded output
	outbuf  [bufferSize]byte
}

func NewUtf8Decoder(r io.Reader) io.Reader {
	return &utf8decoder{r: r}
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
			ndst, nsrc, d.err = Utf8decodeBytes(d.outbuf[0:], d.buf[0:d.nbuf])
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

package bytes

import (
	"bytes"
)

type StringBuffer struct {
	*bytes.Buffer
}

func NewStringBuffer(buf []byte) *StringBuffer {
	return &StringBuffer{
		Buffer: bytes.NewBuffer(buf),
	}
}

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer(buf []byte) *Buffer {
	return &Buffer{
		Buffer: bytes.NewBuffer(buf),
	}
}

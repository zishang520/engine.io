package types

import (
	"bytes"
)

type BytesBuffer struct {
	*bytes.Buffer
}

func NewBytesBuffer(buf []byte) *BytesBuffer {
	return &BytesBuffer{bytes.NewBuffer(buf)}
}

func NewBytesBufferString(s string) *BytesBuffer {
	return &BytesBuffer{bytes.NewBufferString(s)}
}

type StringBuffer struct {
	*bytes.Buffer
}

func NewStringBuffer(buf []byte) *StringBuffer {
	return &StringBuffer{bytes.NewBuffer(buf)}
}

func NewStringBufferString(s string) *StringBuffer {
	return &StringBuffer{bytes.NewBufferString(s)}
}

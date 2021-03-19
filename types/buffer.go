package types

import (
	"bytes"
)

// 字符串buffer，继承至bytes.Buffer
type StringBuffer struct {
	*bytes.Buffer
}

func NewStringBuffer(buf []byte) *StringBuffer {
	return &StringBuffer{
		Buffer: bytes.NewBuffer(buf),
	}
}

// 就是一个bytes.Buffer,没啥特别的
type Buffer struct {
	*bytes.Buffer
}

func NewBuffer(buf []byte) *Buffer {
	return &Buffer{
		Buffer: bytes.NewBuffer(buf),
	}
}

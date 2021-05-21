package types

import (
	"bytes"
	"io"
)

type PacketBuffer interface {
	io.ReadWriter
	io.ReaderFrom
	io.WriterTo
	io.ByteScanner
	io.ByteWriter
	io.RuneScanner
	io.StringWriter
	WriteRune(rune) (int, error)
	Bytes() []byte
	String() string
	Len() int
	Cap() int
	Truncate(int)
	Reset()
	Grow(int)
	Next(int) []byte
	ReadBytes(byte) ([]byte, error)
	ReadString(byte) (string, error)
}

// 字节buffer
type BytesBuffer struct {
	*bytes.Buffer
}

func NewBytesBuffer(buf []byte) PacketBuffer {
	return &BytesBuffer{bytes.NewBuffer(buf)}
}

func NewBytesBufferString(s string) PacketBuffer {
	return &BytesBuffer{bytes.NewBufferString(s)}
}

// 字符串buffer
type StringBuffer struct {
	*bytes.Buffer
}

func NewStringBuffer(buf []byte) PacketBuffer {
	return &StringBuffer{bytes.NewBuffer(buf)}
}

func NewStringBufferString(s string) PacketBuffer {
	return &StringBuffer{bytes.NewBufferString(s)}
}

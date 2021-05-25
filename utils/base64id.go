package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"github.com/zishang520/engine.io/types"
	"sync/atomic"
)

type base64Id struct {
	sequenceNumber uint64
}

var Base64Id = base64Id{0}

func (b *base64Id) GenerateId(ctx *types.HttpContext) (string, error) {
	r := make([]byte, 18)
	if _, err := rand.Read(r[:10]); err != nil {
		return "", err
	}
	if ctx != nil {
		binary.BigEndian.PutUint64(r[10:], ctx.ID())
	} else {
		binary.BigEndian.PutUint64(r[10:], b.sequenceNumber)
		atomic.AddUint64(&b.sequenceNumber, 1)
	}
	return base64.StdEncoding.EncodeToString(r), nil
}

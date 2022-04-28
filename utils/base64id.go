package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"sync/atomic"
)

type base64Id struct {
	sequenceNumber uint64
}

var bid = &base64Id{0}

func Base64Id() *base64Id {
	return bid
}

func (b *base64Id) GenerateId() (string, error) {
	r := make([]byte, 18)
	if _, err := rand.Read(r[:10]); err != nil {
		return "", err
	}
	binary.BigEndian.PutUint64(r[10:], b.sequenceNumber)
	atomic.AddUint64(&b.sequenceNumber, 1)
	return strings.ReplaceAll(strings.ReplaceAll(base64.StdEncoding.EncodeToString(r), "/", "_"), "+", "-"), nil
}

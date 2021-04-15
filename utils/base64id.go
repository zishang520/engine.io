package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
)

type base64Id struct {
	sequenceNumber uint64
}

var Base64Id = base64Id{0}

func (b *base64Id) GenerateId() (string, error) {
	r := make([]byte, 18)
	if _, err := rand.Read(r[:10]); err != nil {
		return "", err
	}
	binary.BigEndian.PutUint64(r[10:], b.sequenceNumber)
	b.sequenceNumber = b.sequenceNumber + 1
	return base64.StdEncoding.EncodeToString(r), nil
}

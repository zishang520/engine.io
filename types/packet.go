package types

import (
	"io"
)

type Packet struct {
	Type string    `json:"type"`
	Data io.Reader `json:"data"`
}

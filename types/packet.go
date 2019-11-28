package types

import (
	"io"
)

type Packet struct {
	Type string
	Data io.Reader
}

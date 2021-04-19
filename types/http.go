package types

import (
	"context"
	"net/http"
)

type HttpContext struct {
	Request  *http.Request
	Response http.ResponseWriter

	ctx context.Context
}

type HttpCompression struct {
	Threshold int
}

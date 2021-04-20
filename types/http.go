package types

import (
	"context"
	"net/http"
)

type HttpContext struct {
	Request  *http.Request
	Response http.ResponseWriter

	Ctx context.Context

	Cleanup Fn
}

type HttpCompression struct {
	Threshold int
}

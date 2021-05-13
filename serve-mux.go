package engineio

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
)

type HttpServer struct {
	*http.Server
	DefaultHandler http.Handler
}

func createServer(addr string, defaultHandler http.Handler) *HttpServer {
	return &HttpServer{Server: &http.Server{Addr: addr, Handler: defaultHandler}, DefaultHandler: defaultHandler}
}

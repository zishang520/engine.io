package types

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/zishang520/engine.io/errors"
	"github.com/zishang520/engine.io/events"
	"github.com/zishang520/engine.io/utils"
)

type HttpContext struct {
	events.EventEmitter

	Websocket *WebSocketConn

	request  *http.Request
	response http.ResponseWriter

	headers *http.Header
	query   *utils.ParameterBag

	method      string
	pathInfo    string
	isHostValid bool

	ctx context.Context

	Cleanup Callable

	isDone bool
	done   chan struct{}
	mu     sync.RWMutex

	wroteHeader bool
	mu_wh       sync.RWMutex
}

func NewHttpContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	c := &HttpContext{}
	c.EventEmitter = events.New()
	c.ctx = r.Context()
	c.done = make(chan struct{})

	c.request = r
	c.response = w

	c.headers = &r.Header
	c.query = utils.NewParameterBag(r.URL.Query())

	c.isHostValid = true

	gone := w.(http.CloseNotifier).CloseNotify()
	go func() {
		select {
		case <-c.ctx.Done():
			c.Close()
			c.Emit("close")
		case <-gone:
			c.Close()
			c.Emit("close")
		}
	}()
	return c
}

func (c *HttpContext) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isDone {
		close(c.done)
		c.isDone = true
	}
}

func (c *HttpContext) Done() <-chan struct{} {
	return c.done
}

func (c *HttpContext) IsWroteHeader() bool {
	c.mu_wh.RLock()
	defer c.mu_wh.RUnlock()

	return c.wroteHeader
}

func (c *HttpContext) IsDone() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.isDone
}

func (c *HttpContext) SetStatusCode(statusCode int) {
	if !c.IsWroteHeader() && !c.IsDone() {
		c.mu_wh.Lock()
		defer c.mu_wh.Unlock()

		c.response.WriteHeader(statusCode)
		c.wroteHeader = true
	}
}

func (c *HttpContext) Write(wb []byte) (int, error) {
	if !c.IsDone() {
		defer c.Close()
		return c.response.Write(wb)
	}
	return 0, errors.New("You cannot write data repeatedly.").Err()
}

func (c *HttpContext) Request() *http.Request {
	return c.request
}

func (c *HttpContext) Response() http.ResponseWriter {
	return c.response
}

func (c *HttpContext) Headers() *http.Header {
	return c.headers
}

func (c *HttpContext) Query() *utils.ParameterBag {
	return c.query
}

func (c *HttpContext) Context() context.Context {
	return c.ctx
}

func (c *HttpContext) GetPathInfo() string {
	if c.pathInfo == "" {
		c.pathInfo = c.request.URL.Path
	}

	return c.pathInfo
}

func (c *HttpContext) Get(key string, _default ...string) string {
	_default = append(_default, "")

	if v, ok := c.query.Get(key); ok {
		return v
	}

	return _default[0]
}

func (c *HttpContext) Gets(key string, _default ...[]string) []string {
	_default = append(_default, []string{})

	if v, ok := c.query.Gets(key); ok {
		return v
	}

	return _default[0]
}

func (c *HttpContext) Method() string {
	return c.GetMethod()
}

func (c *HttpContext) GetMethod() string {
	if c.method != "" {
		return c.method
	}

	c.method = strings.ToUpper(c.request.Method)
	return c.method
}

func (c *HttpContext) Path() string {
	if pattern := strings.Trim(c.GetPathInfo(), "/"); pattern != "" {
		return pattern
	}
	return "/"
}

func (c *HttpContext) GetHost() (string, error) {
	host := c.request.Host
	// trim and remove port number from host
	// host is lowercase as per RFC 952/2181
	host = regexp.MustCompile(`:\d+$`).ReplaceAllString(strings.TrimSpace(host), "")
	// as the host can come from the user (HTTP_HOST and depending on the configuration, SERVER_NAME too can come from the user)
	// check that it does not contain forbidden characters (see RFC 952 and RFC 2181)
	if host != "" {
		if host = regexp.MustCompile(`(?:^\[)?[a-zA-Z0-9-:\]_]+\.?`).ReplaceAllString(host, ""); host != "" {
			if !c.isHostValid {
				return "", nil
			}
			c.isHostValid = false
			return "", errors.New(fmt.Sprintf(`Invalid Host "%s".`, host)).Err()
		}
	}

	return host, nil
}

func (c *HttpContext) UserAgent() string {
	return c.headers.Get("User-Agent")
}

func (c *HttpContext) Secure() bool {
	return c.request.TLS != nil
}

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
	"github.com/zishang520/engine.io/webtransport"
)

type HttpContext struct {
	events.EventEmitter

	Websocket    *WebSocketConn
	WebTransport *webtransport.Conn

	Cleanup Callable

	request  *http.Request
	response http.ResponseWriter

	headers *utils.ParameterBag
	query   *utils.ParameterBag

	method      string
	pathInfo    string
	isHostValid bool

	ctx context.Context

	isDone bool
	done   chan Void
	mu     sync.RWMutex

	statusCode      int
	mu_wh           sync.RWMutex
	ResponseHeaders *utils.ParameterBag

	mu_w sync.Mutex
}

func NewHttpContext(w http.ResponseWriter, r *http.Request) *HttpContext {
	c := &HttpContext{}
	c.EventEmitter = events.New()
	c.ctx = r.Context()
	c.done = make(chan Void)

	c.request = r
	c.response = w

	c.headers = utils.NewParameterBag(r.Header)
	c.query = utils.NewParameterBag(r.URL.Query())

	c.isHostValid = true

	c.ResponseHeaders = utils.NewParameterBag(nil)
	c.ResponseHeaders.With(c.response.Header())

	go func() {
		select {
		case <-c.ctx.Done():
			c.Flush()
			c.Emit("close")
		case <-c.done:
			c.Emit("close")
		}
	}()
	return c
}

func (c *HttpContext) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isDone {
		close(c.done)
		c.isDone = true
	}
}

func (c *HttpContext) Done() <-chan Void {
	return c.done
}

func (c *HttpContext) IsDone() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.isDone
}

func (c *HttpContext) SetStatusCode(statusCode int) {
	c.mu_wh.Lock()
	defer c.mu_wh.Unlock()

	c.statusCode = statusCode
}

func (c *HttpContext) GetStatusCode() int {
	c.mu_wh.RLock()
	defer c.mu_wh.RUnlock()

	return c.statusCode
}

func (c *HttpContext) Write(wb []byte) (int, error) {
	c.mu_w.Lock()
	defer c.mu_w.Unlock()

	if !c.IsDone() {
		defer c.Flush()

		for k, v := range c.ResponseHeaders.All() {
			c.response.Header().Set(k, v[0])
		}
		c.response.WriteHeader(c.GetStatusCode())

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

func (c *HttpContext) Headers() *utils.ParameterBag {
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
	v, _ := c.query.Get(key, _default...)
	return v
}

func (c *HttpContext) Gets(key string, _default ...[]string) []string {
	v, _ := c.query.Gets(key, _default...)
	return v
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
	return c.headers.Peek("User-Agent")
}

func (c *HttpContext) Secure() bool {
	return c.request.TLS != nil
}

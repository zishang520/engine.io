package types

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type Cors struct {
	Origin               interface{} `json:"origin,omitempty"`
	Methods              interface{} `json:"methods,omitempty"`
	AllowedHeaders       interface{} `json:"allowedHeaders,omitempty"`
	Headers              interface{} `json:"headers,omitempty"`
	ExposedHeaders       interface{} `json:"exposedHeaders,omitempty"`
	MaxAge               string      `json:"maxAge,omitempty"`
	Credentials          bool        `json:"credentials,omitempty"`
	PreflightContinue    bool        `json:"preflightContinue,omitempty"`
	OptionsSuccessStatus int         `json:"optionsSuccessStatus,omitempty"`
}

type cors struct {
	options *Cors
	ctx     *HttpContext
	headers []*Kv
	varys   []string
	mu      sync.RWMutex
}

func (c *cors) isOriginAllowed(origin string, allowedOrigin interface{}) bool {
	switch v := allowedOrigin.(type) {
	case []interface{}:
		for _, value := range v {
			if c.isOriginAllowed(origin, value) {
				return true
			}
		}
	case string:
		return origin == v
	case *regexp.Regexp:
		return v.MatchString(origin)
	case bool:
		return v
	}
	return false
}

func (c *cors) configureOrigin() *cors {
	c.mu.Lock()
	defer c.mu.Unlock()

	requestOrigin := c.ctx.Headers().Get("Origin")
	if o, ok := c.options.Origin.(string); ok {
		if o == "*" {
			// allow any origin
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "*",
			})
		} else {
			// fixed origin
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: o,
			})
			c.varys = append(c.varys, "Origin")
		}
	} else {
		// reflect origin
		if c.isOriginAllowed(requestOrigin, c.options.Origin) {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: requestOrigin,
			})
		} else {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "false",
			})
		}
		c.varys = append(c.varys, "Origin")
	}
	return c
}

func (c *cors) configureMethods() *cors {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch methods := c.options.Methods.(type) {
	case string:
		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: methods,
		})
	case []string:
		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: strings.Join(methods, ","),
		})
	}
	return c
}

func (c *cors) configureCredentials() *cors {
	if c.options.Credentials {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Allow-Credentials",
			Value: "true",
		})
	}
	return c
}

func (c *cors) configureAllowedHeaders() *cors {
	allowedHeaders := c.options.AllowedHeaders
	if allowedHeaders == nil {
		allowedHeaders = c.options.Headers
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	switch h := allowedHeaders.(type) {
	case nil:
		// .c.headers wasn't specified, so reflect the request c.headers
		if head := c.ctx.Headers().Get("Access-Control-Request-Headers"); head != "" {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: head,
			})
			c.varys = append(c.varys, "Access-Control-Request-Headers")
		}
	case string:
		if len(h) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: h,
			})
		}
	case []string:
		if len(h) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: strings.Join(h, ","),
			})
		}
	}
	return c
}

func (c *cors) configureExposedHeaders() *cors {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch headers := c.options.ExposedHeaders.(type) {
	case string:
		if len(headers) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: headers,
			})
		}
	case []string:
		if len(headers) > 0 {
			c.headers = append(c.headers, &Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: strings.Join(headers, ","),
			})
		}
	}
	return c
}

func (c *cors) configureMaxAge() *cors {
	if c.options.MaxAge != "" {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.headers = append(c.headers, &Kv{
			Key:   "Access-Control-Expose-Headers",
			Value: c.options.MaxAge,
		})
	}
	return c
}

func parseVary(vary string) *Set[string] {
	end := 0
	start := 0
	list := NewSet[string]()

	// gather tokens
	for i, l := 0, len(vary); i < l; i++ {
		switch vary[i] {
		case ' ': /*   */
			if start == end {
				end = i + 1
				start = end
			}
		case ',': /* , */
			list.Add(vary[start:end])
			end = i + 1
			start = end
		default:
			end = i + 1
		}
	}

	if end := vary[start:end]; len(end) > 0 {
		// final token
		list.Add(end)
	}

	return list
}

func (c *cors) applyHeaders() *cors {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, header := range c.headers {
		c.ctx.Response().Header().Set(header.Key, header.Value)
	}
	vary := c.ctx.Response().Header().Get("Vary")
	if vary == "*" {
		c.ctx.Response().Header().Set("Vary", "*")
	} else {
		if len(c.varys) > 0 {
			varys := parseVary(vary)
			varys.Add(c.varys...)
			c.ctx.Response().Header().Set("Vary", strings.Join(varys.Keys(), ", "))
		}
	}
	return c
}

func corsFunc(options *Cors, ctx *HttpContext, next Callable) {
	c := &cors{
		options: options,
		ctx:     ctx,
		headers: []*Kv{},
	}
	method := c.ctx.Method()

	if http.MethodOptions == method {
		// preflight
		c.configureOrigin().configureCredentials().configureMethods().configureAllowedHeaders().configureMaxAge().configureExposedHeaders().applyHeaders()
		if options.PreflightContinue {
			next()
		} else {
			// Safari (and potentially other browsers) need content-length 0,
			//   for 204 or they just hang waiting for a body
			ctx.Response().Header().Set("Content-Length", "0")
			ctx.SetStatusCode(options.OptionsSuccessStatus)
			ctx.Write(nil)
		}
	} else {
		// actual response
		c.configureOrigin().configureCredentials().configureExposedHeaders().applyHeaders()
		next()
	}
}

var defaults = &Cors{
	Origin:               `*`,
	Methods:              `GET,HEAD,PUT,PATCH,POST,DELETE`,
	PreflightContinue:    false,
	OptionsSuccessStatus: 204,
}

func MiddlewareWrapper(options *Cors) func(*HttpContext, Callable) {
	if options == nil {
		options = defaults
	} else {
		if options.Origin == nil {
			options.Origin = "*"
		}

		if options.Methods == nil {
			options.Methods = `GET,HEAD,PUT,PATCH,POST,DELETE`
		}

		if options.OptionsSuccessStatus == 0 {
			options.OptionsSuccessStatus = http.StatusNoContent
		}
	}

	return func(ctx *HttpContext, next Callable) {
		if options.Origin == nil {
			next()
		} else {
			corsFunc(options, ctx, next)
		}
	}
}

package types

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type (
	Cors struct {
		Origin               any    `json:"origin,omitempty" mapstructure:"origin,omitempty" msgpack:"origin,omitempty"`
		Methods              any    `json:"methods,omitempty" mapstructure:"methods,omitempty" msgpack:"methods,omitempty"`
		AllowedHeaders       any    `json:"allowedHeaders,omitempty" mapstructure:"allowedHeaders,omitempty" msgpack:"allowedHeaders,omitempty"`
		Headers              any    `json:"headers,omitempty" mapstructure:"headers,omitempty" msgpack:"headers,omitempty"`
		ExposedHeaders       any    `json:"exposedHeaders,omitempty" mapstructure:"exposedHeaders,omitempty" msgpack:"exposedHeaders,omitempty"`
		MaxAge               string `json:"maxAge,omitempty" mapstructure:"maxAge,omitempty" msgpack:"maxAge,omitempty"`
		Credentials          bool   `json:"credentials,omitempty" mapstructure:"credentials,omitempty" msgpack:"credentials,omitempty"`
		PreflightContinue    bool   `json:"preflightContinue,omitempty" mapstructure:"preflightContinue,omitempty" msgpack:"preflightContinue,omitempty"`
		OptionsSuccessStatus int    `json:"optionsSuccessStatus,omitempty" mapstructure:"optionsSuccessStatus,omitempty" msgpack:"optionsSuccessStatus,omitempty"`
	}

	Kv struct {
		Key   string
		Value string
	}

	cors struct {
		options *Cors
		ctx     *HttpContext
		headers []*Kv
		varys   []string
		mu      sync.RWMutex
	}
)

func (c *cors) isOriginAllowed(origin string, allowedOrigin any) bool {
	switch v := allowedOrigin.(type) {
	case []any:
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

	requestOrigin := c.ctx.Headers().Peek("Origin")
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
		if head := c.ctx.Headers().Peek("Access-Control-Request-Headers"); head != "" {
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
		c.ctx.ResponseHeaders.Set(header.Key, header.Value)
	}
	if vary := c.ctx.ResponseHeaders.Peek("Vary"); vary == "*" {
		c.ctx.ResponseHeaders.Set("Vary", "*")
	} else {
		if len(c.varys) > 0 {
			varys := parseVary(vary)
			varys.Add(c.varys...)
			c.ctx.ResponseHeaders.Set("Vary", strings.Join(varys.Keys(), ", "))
		}
	}
	return c
}

func corsMiddleware(options *Cors, ctx *HttpContext, next func(error)) {
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
			next(nil)
		} else {
			// Safari (and potentially other browsers) need content-length 0,
			//   for 204 or they just hang waiting for a body
			ctx.ResponseHeaders.Set("Content-Length", "0")
			ctx.SetStatusCode(options.OptionsSuccessStatus)
			ctx.Write(nil)
		}
	} else {
		// actual response
		c.configureOrigin().configureCredentials().configureExposedHeaders().applyHeaders()
		next(nil)
	}
}

var defaultCors = &Cors{
	Origin:               `*`,
	Methods:              `GET,HEAD,PUT,PATCH,POST,DELETE`,
	PreflightContinue:    false,
	OptionsSuccessStatus: 204,
}

func MiddlewareWrapper(options *Cors) func(*HttpContext, func(error)) {
	if options == nil {
		options = defaultCors
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

	return func(ctx *HttpContext, next func(error)) {
		if options.Origin == nil {
			next(nil)
		} else {
			corsMiddleware(options, ctx, next)
		}
	}
}

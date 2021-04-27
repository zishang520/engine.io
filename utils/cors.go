package utils

import (
	"github.com/zishang520/engine.io/types"
	"strings"
)

type cors struct {
	options *types.Cors
	ctx     *types.HttpContext
	headers []*types.Kv
}

var initCors = &Cors{
	Origin:               `*`,
	Methods:              `GET,HEAD,PUT,PATCH,POST,DELETE`,
	PreflightContinue:    false,
	OptionsSuccessStatus: 204,
}

func (c *cors) isOriginAllowed(origin string, allowedOrigin interface{}) bool {
	switch v := allowedOrigin.(type) {
	case []interface{}:
		for _, a := range v {
			if c.isOriginAllowed(origin, a) {
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
	requestOrigin := c.ctx.Request.Header.Get("Origin")
	if o, ok := c.options.Origin.(string); ok {
		if o == "*" {
			// allow any origin
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "*",
			})
		} else {
			// fixed origin
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: o,
			})
		}
	} else {
		// reflect origin
		if c.isOriginAllowed(requestOrigin, c.options.Origin) {
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: requestOrigin,
			})
		} else {
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "false",
			})
		}
	}
	return c
}

func (c *cors) configureMethods() *cors {
	switch methods := c.options.Methods.(type) {
	case string:
		c.headers = append(c.headers, &types.Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: methods,
		})
	case []string:
		c.headers = append(c.headers, &types.Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: strings.Join(methods, ","),
		})
	}
	return c
}

func (c *cors) configureCredentials() *cors {
	if c.options.Credentials {
		c.headers = append(c.headers, &types.Kv{
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

	switch h := allowedHeaders.(type) {
	case nil:
		head := c.ctx.Request.Header.Get("Access-Control-Request-Headers") // .c.headers wasn't specified, so reflect the request c.headers
		if head != "" {
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Request-Headers",
				Value: head,
			})
		}
	case string:
		if headers != "" {
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: h,
			})
		}
	case []string:
		if len(headers) > 0 {
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: strings.Join(methods, ","),
			})
		}
	}
	return c
}

func (c *cors) configureExposedHeaders() *cors {
	switch headers := c.options.ExposedHeaders.(type) {
	case string:
		if headers != "" {
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: methods,
			})
		}
	case []string:
		if len(headers) > 0 {
			c.headers = append(c.headers, &types.Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: strings.Join(methods, ","),
			})
		}
	}
	return c
}

func (c *cors) configureMaxAge() *cors {
	if c.options.MaxAge != "" {
		c.headers = append(c.headers, &types.Kv{
			Key:   "Access-Control-Expose-Headers",
			Value: c.options.MaxAge,
		})
	}
	return c
}

func (c *cors) applyHeaders() *cors {
	for _, header := range c.headers {
		c.ctx.Response.Header().Set(header.Key, c.headers.Value)
	}
}

func _cors(options *types.Cors, ctx *types.HttpContext, next types.Fn) {
	c := &_cors{
		options: *types.Cors,
		ctx:     *types.HttpContext,
		headers: []*types.Kv{},
	}
	method = strings.ToUpper(c.ctx.Request.Method)

	if method == "OPTIONS" {
		// preflight
		c.configureOrigin().configureCredentials().configureMethods().configureAllowedHeaders().configureMaxAge().configureExposedHeaders().applyHeaders()
		if options.PreflightContinue {
			next()
		} else {
			// Safari (and potentially other browsers) need content-length 0,
			//   for 204 or they just hang waiting for a body
			ctx.Response.WriteHeader(options.OptionsSuccessStatus)
			ctx.Response.Header().Set("Content-Length", "0")
			ctx.Response.Write(nil)
		}
	} else {
		// actual response
		c.configureOrigin().configureCredentials().configureExposedHeaders().applyHeaders()
		next()
	}
}

func MiddlewareWrapper(options *types.Cors) func(*types.HttpContext, types.Fn) {
	return func(ctx *types.HttpContext, next types.Fn) {
		corsOptions = initCors.Assign(options)
		if corsOptions.Origin == nil {
			next()
		} else {
			cors(corsOptions, ctx, next)
		}
	}
}

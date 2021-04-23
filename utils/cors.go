package utils

import (
	"github.com/zishang520/engine.io/types"
	"strings"
)

type headers []*types.Kv

var initCors = &Cors{
	Origin:               `*`,
	Methods:              `GET,HEAD,PUT,PATCH,POST,DELETE`,
	PreflightContinue:    false,
	OptionsSuccessStatus: 204,
}

func isOriginAllowed(origin string, allowedOrigin interface{}) bool {
	switch ao := allowedOrigin.(type) {
	case []interface{}:
		for _, a := range ao {
			if isOriginAllowed(origin, a) {
				return true
			}
		}
	case string:
		return origin == ao
	case *regexp.Regexp:
		return ao.MatchString(origin)
	case bool:
		return ao
	}
	return false
}

func configureOrigin(options *types.Cors, ctx *types.HttpContext) (headers []*types.Kv) {
	requestOrigin := ctx.Request.Header.Get("Origin")
	// ctx.Response.Header().Add(key, value)
	if o, ok := options.Origin.(string); ok {
		if o == "*" {
			// allow any origin
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "*",
			})
		} else {
			// fixed origin
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: o,
			})
		}
	} else {
		// reflect origin
		if isOriginAllowed(requestOrigin, options.Origin) {
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: requestOrigin,
			})
		} else {
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "false",
			})
		}
	}
	return headers
}

func configureMethods(options *types.Cors) (headers []*types.Kv) {
	switch methods := options.Methods.(type) {
	case string:
		headers = append(headers, &types.Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: methods,
		})
	case []string:
		headers = append(headers, &types.Kv{
			Key:   "Access-Control-Allow-Methods",
			Value: strings.Join(methods, ","),
		})
	}
	return headers
}

func configureCredentials(options *types.Cors) (headers []*types.Kv) {
	if options.Credentials {
		headers = append(headers, &types.Kv{
			Key:   "Access-Control-Allow-Credentials",
			Value: "true",
		})
	}
	return headers
}

func configureAllowedHeaders(options *types.Cors, ctx *types.HttpContext) (headers []*types.Kv) {
	allowedHeaders := options.AllowedHeaders
	if allowedHeaders == nil {
		allowedHeaders = options.Headers
	}

	switch h := allowedHeaders.(type) {
	case nil:
		head := ctx.Request.Header.Get("Access-Control-Request-Headers") // .headers wasn't specified, so reflect the request headers
		if head != "" {
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Request-Headers",
				Value: head,
			})
		}
	case string:
		if headers != "" {
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: h,
			})
		}
	case []string:
		if len(headers) > 0 {
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Allow-Headers",
				Value: strings.Join(methods, ","),
			})
		}
	}
	return headers
}

func configureExposedHeaders(options *types.Cors) (headers []*types.Kv) {
	switch headers := options.ExposedHeaders.(type) {
	case string:
		if headers != "" {
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: methods,
			})
		}
	case []string:
		if len(headers) > 0 {
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Expose-Headers",
				Value: strings.Join(methods, ","),
			})
		}
	}
	return headers
}

func configureMaxAge(options *types.Cors) (headers []*types.Kv) {
	if options.MaxAge != "" {
		headers = append(headers, &types.Kv{
			Key:   "Access-Control-Expose-Headers",
			Value: options.MaxAge,
		})
	}
	return headers
}

func applyHeaders(headers []*types.Kv, ctx *types.HttpContext) {
	for _, header := range headers {
		ctx.Response.Header().Set(header.Key, headers.Value)
	}
}

func cors(options *types.Cors, ctx *types.HttpContext, next types.Fn) {
	headers := []*types.Kv{}
	method = strings.ToUpper(ctx.Request.Method)

	if method == "OPTIONS" {
		// preflight
		headers = append(headers, configureOrigin(options, ctx)...)
		headers = append(headers, configureCredentials(options)...)
		headers = append(headers, configureMethods(options)...)
		headers = append(headers, configureAllowedHeaders(options, ctx)...)
		headers = append(headers, configureMaxAge(options)...)
		headers = append(headers, configureExposedHeaders(options)...)
		applyHeaders(headers, ctx)

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
		headers = append(headers, configureOrigin(options, ctx)...)
		headers = append(headers, configureCredentials(options)...)
		headers = append(headers, configureExposedHeaders(options)...)
		applyHeaders(headers, ctx)
		next()
	}
}

func MiddlewareWrapper(options *types.Cors) {
	return func(ctx *types.HttpContext, next types.Fn) {
		corsOptions = initCors.Assign(options)
		if corsOptions.Origin == nil {
			next()
		} else {
			cors(corsOptions, ctx, next)
		}
	}
}

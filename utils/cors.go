package utils

import (
	"github.com/zishang520/engine.io/types"
)

type headers []*types.Kv

var defaults = map[string]string{
	"Origin":               `*`,
	"Methods":              `GET,HEAD,PUT,PATCH,POST,DELETE`,
	"PreflightContinue":    false,
	"OptionsSuccessStatus": 204,
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
	if o, ok := options.Origin.(type); ok {
		if o == "*" {
			// allow any origin
			headers = append(headers, &types.Kv{
				Key:   "Access-Control-Allow-Origin",
				Value: "*",
			})
		} else {
			// fixed origin
			headers = append(headers, []*types.Kv{
				&types.Kv{
					Key:   "Access-Control-Allow-Origin",
					Value: o,
				},
				&types.Kv{
					Key:   "Vary",
					Value: "Origin",
				},
			}...)
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
		headers = append(headers, &types.Kv{
			Key:   "Vary",
			Value: "Origin",
		})
	}
	return headers
}

func configureMethods(*types.Cors) (headers []*types.Kv) {
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

func Cors() {

}

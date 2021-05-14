package utils

import (
	"net"
	"strings"
)

func Contains(haystack string, needles []string) string {

	for _, needle := range needles {
		if needle != "" && strings.Index(haystack, needle) > -1 {
			return needle
		}
	}

	return ""
}

func StripHostPort(h string) string {
	if strings.IndexByte(h, ':') == -1 {
		return h
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		return h
	}
	return host
}

func CleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

/**
 * From https://github.com/nodejs/node/blob/v8.4.0/lib/_http_common.js#L303-L354
 *
 * True if val contains an invalid field-vchar
 *  field-value    = *( field-content / obs-fold )
 *  field-content  = field-vchar [ 1*( SP / HTAB ) field-vchar ]
 *  field-vchar    = VCHAR / obs-text
 *
 * checkInvalidHeaderChar() is currently designed to be inlinable by v8,
 * so take care when making changes to the implementation so that the source
 * code size does not exceed v8's default max_inlined_source_size setting.
 **/
var validHdrChars = [...]bool{
	false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, // 0 - 15
	false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, // 16 - 31
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 32 - 47
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 48 - 63
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 64 - 79
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 80 - 95
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 96 - 111
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, // 112 - 127
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // 128 ...
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
	true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, // ... 255
}

func CheckInvalidHeaderChar(val []byte) bool {
	length := len(val)
	if length < 1 {
		return false
	}
	if !validHdrChars[val[0]] {
		// debug(`invalid header, index 0, char "%s"`, val.charCodeAt(0))
		return true
	}
	if length < 2 {
		return false
	}
	if !validHdrChars[val[1]] {
		// debug(`invalid header, index true, char "%s"`, val.charCodeAt(1))
		return true
	}
	if length < 3 {
		return false
	}
	if !validHdrChars[val[2]] {
		// debug(`invalid header, index 2, char "%s"`, val.charCodeAt(2))
		return true
	}
	if length < 4 {
		return false
	}
	if !validHdrChars[val[3]] {
		// debug(`invalid header, index 3, char "%s"`, val.charCodeAt(3))
		return true
	}
	for i := 4; i < length; i += 1 {
		if !validHdrChars[val[i]] {
			// debug(`invalid header, index "%i", char "%s"`, i, val.charCodeAt(i))
			return true
		}
	}
	return false
}

package utils

import (
	"net"
	"path"
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
	if pl := len(p); p[pl-1] == '/' && np != "/" {
		if pl == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}

// isLWS reports whether b is linear white space, according
// to http://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2
//      LWS            = [CRLF] 1*( SP | HT )
func isLWS(b byte) bool { return b == ' ' || b == '\t' }

// isCTL reports whether b is a control byte, according
// to http://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2
//      CTL            = <any US-ASCII control character
//                       (octets 0 - 31) and DEL (127)>
func isCTL(b byte) bool {
	return b < ' ' || b == 0x7f // a CTL
}

// CheckInvalidHeaderChar reports whether v is an invalid "field-value" according to
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2 :
//
//        message-header = field-name ":" [ field-value ]
//        field-value    = *( field-content | LWS )
//        field-content  = <the OCTETs making up the field-value
//                         and consisting of either *TEXT or combinations
//                         of token, separators, and quoted-string>
//
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec2.html#sec2.2 :
//
//        TEXT           = <any OCTET except CTLs,
//                          but including LWS>
//        LWS            = [CRLF] 1*( SP | HT )
//        CTL            = <any US-ASCII control character
//                         (octets 0 - 31) and DEL (127)>
//
// RFC 7230 says:
//  field-value    = *( field-content / obs-fold )
//  obj-fold       =  N/A to http2, and deprecated
//  field-content  = field-vchar [ 1*( SP / HTAB ) field-vchar ]
//  field-vchar    = VCHAR / obs-text
//  obs-text       = %x80-FF
//  VCHAR          = "any visible [USASCII] character"
//
// http2 further says: "Similarly, HTTP/2 allows header field values
// that are not valid. While most of the values that can be encoded
// will not alter header field parsing, carriage return (CR, ASCII
// 0xd), line feed (LF, ASCII 0xa), and the zero character (NUL, ASCII
// 0x0) might be exploited by an attacker if they are translated
// verbatim. Any request or response that contains a character not
// permitted in a header field value MUST be treated as malformed
// (Section 8.1.2.6). Valid characters are defined by the
// field-content ABNF rule in Section 3.2 of [RFC7230]."
//
// This function does not (yet?) properly handle the rejection of
// strings that begin or end with SP or HTAB.
func CheckInvalidHeaderChar(val string) bool {
	for i, l := 0, len(val); i < l; i++ {
		b := val[i]
		if isCTL(b) && !isLWS(b) {
			return true
		}
	}
	return false
}
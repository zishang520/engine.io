package utils

import (
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

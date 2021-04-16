package utils

import (
	"strings"
)

func Contains(haystack string, needles []string) bool {

	for _, needle := range needles {
		if needle != "" && strings.Index(haystack, needle) > 0 {
			return true
		}
	}

	return false
}

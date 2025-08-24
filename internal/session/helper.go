package session

import (
	"regexp"
	"strings"
)

func matchString(str, pattern string) bool {
	if pattern == "" {
		return true
	}

	if isRegex(pattern) {

		return matchRegex(str, pattern)
	}

	return strings.Contains(strings.ToLower(str), strings.ToLower(pattern))
}

func isRegex(query string) bool {
	return len(query) > 2 && strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/")
}

func matchRegex(target, regexPattern string) bool {
	pattern := regexPattern[1 : len(regexPattern)-1]

	re, err := regexp.Compile(pattern)
	if err != nil {
		return strings.Contains(strings.ToLower(target), strings.ToLower(regexPattern))
	}

	return re.MatchString(target)
}

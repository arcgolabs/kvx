package shared

import "strings"

func matchesPattern(key, pattern string) bool {
	prefix, hasWildcard := strings.CutSuffix(pattern, "*")
	if hasWildcard {
		return strings.HasPrefix(key, prefix)
	}

	return key == pattern
}

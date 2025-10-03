package main

import (
	"regexp"
	"strings"

	"github.com/concourse/concourse/atc"
)

var (
	placeholderRegex = regexp.MustCompile(`{([^}]+)}`)
)

// ExpandPattern tries to expand a pattern with curly-braced placeholders using
// a map for replacements. For example:
//
//	ExpandPattern("{a}-{b}", atc.Source{"a": "foo", "b": "bar", "c": "baz"})
//
// expands to "foo-bar". ExpandPattern returns true only if all placeholders
// were expanded.
func ExpandPattern(pattern string, attrs atc.Source) (string, bool) {
	// Expand all {placeholder} patterns using data from resource config.
	expanded := placeholderRegex.ReplaceAllStringFunc(pattern, func(placeholder string) string {
		key := strings.Trim(placeholder, "{}")
		// Only replace string values
		if val, ok := attrs[key].(string); ok {
			return val
		}
		// Return the placeholder itself if we can't find replacement.
		return placeholder
	})
	if strings.ContainsAny(expanded, "{}") {
		// Not fully expanded, signal failure
		return "", false
	}
	return expanded, true
}

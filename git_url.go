package main

import (
	"regexp"
	"strings"
)

var (
	gitURIRegex = regexp.MustCompile(`(https?://|git://|[^@]+@)(?P<host>[-.a-z0-9]+)[:/](?P<repository>.*)`)
)

func SameGitRepository(url1, url2 string) bool {
	if url1 == url2 {
		return true
	}
	matches1 := gitURIRegex.FindStringSubmatch(url1)
	if matches1 == nil {
		return false
	}
	matches2 := gitURIRegex.FindStringSubmatch(url2)
	if matches2 == nil {
		return false
	}

	return matches1[2] == matches2[2] && strings.TrimSuffix(matches1[3], ".git") == strings.TrimSuffix(matches2[3], ".git")
}

package main

import (
	"fmt"

	"github.com/concourse/concourse/atc"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// The various GitHub-related resource types store their repository
// configuration in different ways. This function handles the different
// approaches and returns a standardised repository URL.
func ConstructGitHubUriFromConfig(resource atc.ResourceConfig) (string, bool) {
	keys := make([]string, 0, len(resource.Source))
	for k := range resource.Source {
		keys = append(keys, k)
	}

	// Config in `uri` key for:
	// - concourse/git-resource
	if contains(keys, "uri") {
		return resource.Source["uri"].(string), true
	}

	// Config in `repository` key only for:
	// - telia-oss/github-pr-resource
	if contains(keys, "repository") {
		return fmt.Sprintf("https://github.com/%s", resource.Source["repository"].(string)), true
	}

	return "", false
}

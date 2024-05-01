package main

import (
	"testing"

	"github.com/concourse/concourse/atc"
)

func TestGitHubUriConstruction(t *testing.T) {
	cases := []struct {
		ResourceConfig atc.ResourceConfig
		Result         string
	}{
		{atc.ResourceConfig{Source: atc.Source{"uri": "https://github.com/some/repo"}}, "https://github.com/some/repo"},
		{atc.ResourceConfig{Source: atc.Source{"repository": "some/repo"}}, "https://github.com/some/repo"},
	}

	for nr, c := range cases {
		uri, ok := ConstructGitHubUriFromConfig(c.ResourceConfig)
		if !ok || uri != c.Result {
			t.Errorf("Test case %d failed.", nr+1)
		}
	}
}

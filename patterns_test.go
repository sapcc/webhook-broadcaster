package main

import (
	"testing"

	"github.com/concourse/concourse/atc"
)

func TestExpand(t *testing.T) {
	cases := []struct {
		name        string
		pattern     string
		attrs       atc.Source
		expectedVal string
		expectedOK  bool
	}{
		{
			// Expansion fails if one attr is missing
			name:        "missing attr value",
			pattern:     "https://g.com/{owner}/{repo}",
			attrs:       atc.Source{"owner": "john"},
			expectedVal: "",
			expectedOK:  false,
		},
		{
			// Expansion fails if we try to substitute non-string values
			name:        "non-string attr values",
			pattern:     "https://g.com/{org}/{repo}",
			attrs:       atc.Source{"org": []string{"bob"}, "repo": 400},
			expectedVal: "",
			expectedOK:  false,
		},
		{
			name:        "simple-https-uri",
			pattern:     "https://github.com/{owner}/{repo}",
			attrs:       atc.Source{"owner": "john", "repo": "doe"},
			expectedVal: "https://github.com/john/doe",
			expectedOK:  true,
		},
		{
			name:        "simple-git-uri",
			pattern:     "git@git.foo:{repo}",
			attrs:       atc.Source{"repo": "jane/doe.git"},
			expectedVal: "git@git.foo:jane/doe.git",
			expectedOK:  true,
		},
		{
			name:        "simple-branch-name",
			pattern:     "{branch}",
			attrs:       atc.Source{"branch": "master"},
			expectedVal: "master",
			expectedOK:  true,
		},
	}

	for _, c := range cases {
		actual, ok := ExpandPattern(c.pattern, c.attrs)
		if ok != c.expectedOK || actual != c.expectedVal {
			t.Errorf("%s: expected %q,%v, was %q,%v.", c.name, c.expectedVal, c.expectedOK, actual, ok)
		}
	}
}

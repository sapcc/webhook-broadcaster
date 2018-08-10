package main

import "testing"

func TestGitURIComparision(t *testing.T) {

	cases := []struct {
		URL1   string
		URL2   string
		Result bool
	}{
		{"https://git.foo/some/repo", "https://git.foo/other/repo", false},
		{"https://git.foo/some/repo", "https://git.foo/some/repo.git", true},
		{"https://git.foo/some/repo.git", "git://git.foo/some/repo", true},
		{"git@git.foo:some/repo.git", "https://git.foo/some/repo.git", true},
		{"git@git.foo:some/repo.git", "nase@git.foo:some/repo.git", true},
		{"git@git.foo:some/repo.git", "nase@git.foo:some/repo2.git", false},
		{"git@git.bar:some/repo.git", "nase@git.foo:some/repo.git", false},
	}

	for nr, c := range cases {
		if SameGitRepository(c.URL1, c.URL2) != c.Result {
			t.Errorf("Test case %d failed.", nr+1)
		}

	}

}

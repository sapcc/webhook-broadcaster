package main

import "testing"

func TestMatchFiles(t *testing.T) {
	cases := []struct {
		patterns []string
		files    []string
		Result   bool
	}{
		{[]string{"ap-ae-1/values/globals.yaml"}, []string{"qa-de-1/values/designate.yaml"}, false},
	}
	for nr, c := range cases {
		if matchFiles(c.patterns, c.files) != c.Result {
			t.Errorf("Test case %d failed.", nr+1)
		}

	}

}

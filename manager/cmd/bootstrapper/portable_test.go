package main

import "testing"

func TestPortableMinimumVersion(t *testing.T) {
	for _, v := range []struct {
		have, want string
		ok         bool
	}{
		{"0.4.2", "0.4.4", false}, {"0.4.4", "0.4.4", true}, {"0.4.5", "0.4.4", true}, {"0.4.10", "0.4.4", true}, {"1.0.0", "0.4.4", true}, {"0.4.4-alpha", "0.4.4", false}, {"0.4", "0.4.4", false}, {"-1.4.4", "0.4.4", false}, {"0.4.4", "bad", false}} {
		if got := versionAtLeast(v.have, v.want); got != v.ok {
			t.Fatal(v, got)
		}
	}
}

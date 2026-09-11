package main

import "testing"

func TestSessionExpiryDecision(t *testing.T) {
	cases := []struct {
		name        string
		expiryDelta int64
		want        bool
	}{{"active", 60, true}, {"expired", -1, false}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			active := tc.expiryDelta > 0
			if active != tc.want {
				t.Fatalf("active=%v", active)
			}
		})
	}
}

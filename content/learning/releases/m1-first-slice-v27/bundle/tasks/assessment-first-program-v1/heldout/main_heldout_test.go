package main

import "testing"

func TestGreetingHandlesDefaultsAndValues(t *testing.T) {
	tests := []struct {
		name     string
		attempts int
		want     string
	}{
		{name: "Wayne", attempts: 3, want: "welcome, Wayne; attempts=3"},
		{name: "", attempts: 0, want: "welcome, Gopher; attempts=0"},
	}
	for _, test := range tests {
		if got := greeting(test.name, test.attempts); got != test.want {
			t.Fatalf("greeting(%q, %d) = %q, want %q", test.name, test.attempts, got, test.want)
		}
	}
}

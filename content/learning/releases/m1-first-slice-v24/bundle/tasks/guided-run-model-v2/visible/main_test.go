package main

import "testing"

func TestWelcomeUsesTheProvidedName(t *testing.T) {
	tests := map[string]string{
		"Gopher": "welcome, Gopher",
		"Wayne":  "welcome, Wayne",
	}
	for name, want := range tests {
		if got := welcome(name); got != want {
			t.Fatalf("welcome(%q) = %q, want %q", name, got, want)
		}
	}
}

package main

import "testing"

func TestGreetingFormatsTheCommandOutput(t *testing.T) {
	if got, want := greeting("Gopher", 1), "welcome, Gopher; attempts=1"; got != want {
		t.Fatalf("greeting() = %q, want %q", got, want)
	}
}

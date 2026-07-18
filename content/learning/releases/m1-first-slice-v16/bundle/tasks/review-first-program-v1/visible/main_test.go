package main

import "testing"

func TestServiceStatusReady(t *testing.T) {
	if got, want := serviceStatus("api", true), "api=ready"; got != want {
		t.Fatalf("serviceStatus() = %q, want %q", got, want)
	}
}

package main

import "testing"

func TestServiceStatusHandlesVariantInputs(t *testing.T) {
	if got, want := serviceStatus("worker", false), "worker=retry"; got != want {
		t.Fatalf("serviceStatus() = %q, want %q", got, want)
	}
}

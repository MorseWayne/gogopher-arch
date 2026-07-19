package checkcfg

import "testing"

func TestNormalizeName(t *testing.T) {
	got, err := NormalizeName(" API ")
	if err != nil || got != "api" {
		t.Fatalf("NormalizeName() = %q, %v", got, err)
	}
}

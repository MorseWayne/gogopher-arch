package config

import "testing"

func TestNormalizeRejectsEmptyTargets(t *testing.T) {
	if _, err := Normalize(Config{}); err == nil {
		t.Fatal("Normalize(Config{}) error = nil")
	}
}

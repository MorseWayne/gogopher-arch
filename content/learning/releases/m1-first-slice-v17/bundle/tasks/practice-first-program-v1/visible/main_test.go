package main

import "testing"

func TestReportUsesNameAndStatus(t *testing.T) {
	tests := []struct {
		name   string
		passed bool
		want   string
	}{
		{name: "build", passed: true, want: "build: ready"},
		{name: "test", passed: false, want: "test: retry"},
	}
	for _, test := range tests {
		if got := report(test.name, test.passed); got != test.want {
			t.Fatalf("report(%q, %v) = %q, want %q", test.name, test.passed, got, test.want)
		}
	}
}

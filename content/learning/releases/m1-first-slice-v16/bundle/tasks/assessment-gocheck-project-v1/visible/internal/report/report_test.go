package report

import (
	"testing"

	"gocheck/internal/check"
)

func TestTextUsesTheStableStatusProtocol(t *testing.T) {
	results := []check.Result{
		{Name: "api", URL: "http://api", StatusCode: 204},
		{Name: "db", URL: "http://db", StatusCode: 503},
		{Name: "cache", URL: "http://cache", Error: "timeout"},
	}
	if got, want := Text(results), "api\tok\t204\ndb\tfail\t503\ncache\terror\t0\n"; got != want {
		t.Fatalf("Text() = %q, want %q", got, want)
	}
}

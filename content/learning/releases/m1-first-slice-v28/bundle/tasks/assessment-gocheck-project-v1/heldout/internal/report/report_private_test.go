package report

import (
	"testing"

	"gocheck/internal/check"
)

func TestReportsAreStableInTextAndJSON(t *testing.T) {
	results := []check.Result{{Name: "api", URL: "http://api", StatusCode: 204}, {Name: "db", URL: "http://db", StatusCode: 503}}
	if got, want := Text(results), "api\tok\t204\ndb\tfail\t503\n"; got != want {
		t.Fatalf("Text() = %q, want %q", got, want)
	}
	got, err := JSON(results)
	if err != nil {
		t.Fatal(err)
	}
	want := "[{\"name\":\"api\",\"url\":\"http://api\",\"status_code\":204},{\"name\":\"db\",\"url\":\"http://db\",\"status_code\":503}]\n"
	if got != want {
		t.Fatalf("JSON() = %q, want %q", got, want)
	}
}

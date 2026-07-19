package webhook

import (
	"net/http"
	"testing"
)

func TestNewValidatesClientConfiguration(t *testing.T) {
	for _, test := range []struct {
		client *http.Client
		base   string
		max    int64
	}{{nil, "https://delivery.example", 1024}, {&http.Client{}, "not a url", 1024}, {&http.Client{}, "https://delivery.example", 0}} {
		if _, err := New(test.client, test.base, test.max); err == nil {
			t.Fatalf("accepted invalid config %#v", test)
		}
	}
	if _, err := New(&http.Client{}, "https://delivery.example", 1024); err != nil {
		t.Fatal(err)
	}
}

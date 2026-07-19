package gocheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckAllAndRenderTextKeepTheProjectContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/bad") {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	targets := []Target{{Name: "api", URL: server.URL + "/ok"}, {Name: "db", URL: server.URL + "/bad"}}
	results, err := CheckAll(context.Background(), server.Client(), targets, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].Name != "api" || results[1].StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("CheckAll() = %#v", results)
	}
	if got, want := RenderText(results), "api\tok\t204\ndb\tfail\t503\n"; got != want {
		t.Fatalf("RenderText() = %q, want %q", got, want)
	}
}

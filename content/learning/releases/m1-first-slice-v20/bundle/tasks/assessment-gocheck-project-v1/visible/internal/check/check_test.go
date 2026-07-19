package check

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllChecksTargetsAndPreservesOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	targets := []Target{{Name: "first", URL: server.URL + "/first"}, {Name: "second", URL: server.URL + "/second"}}
	results, err := All(context.Background(), server.Client(), targets, 2, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].Name != "first" || results[1].Name != "second" || results[0].StatusCode != http.StatusNoContent {
		t.Fatalf("All() = %#v", results)
	}
}

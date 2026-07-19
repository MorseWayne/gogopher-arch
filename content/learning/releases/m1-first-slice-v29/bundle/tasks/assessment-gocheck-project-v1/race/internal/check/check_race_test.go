package check

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllIsRaceFreeUnderRepeatedConcurrentChecks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	targets := make([]Target, 24)
	for index := range targets {
		targets[index] = Target{Name: "target", URL: server.URL}
	}
	for range 5 {
		results, err := All(context.Background(), server.Client(), targets, 4, time.Second)
		if err != nil || len(results) != len(targets) {
			t.Fatalf("All() len=%d error=%v", len(results), err)
		}
	}
}

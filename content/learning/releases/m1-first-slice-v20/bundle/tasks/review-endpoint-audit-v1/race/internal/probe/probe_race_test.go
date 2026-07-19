package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllRemainsRaceFreeAcrossBatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	checks := make([]Check, 20)
	for index := range checks {
		checks[index] = Check{Name: "check", URL: server.URL, AcceptStatus: http.StatusAccepted}
	}
	for range 4 {
		results, err := All(context.Background(), server.Client(), checks, 4, time.Second)
		if err != nil || len(results) != len(checks) {
			t.Fatalf("All() len=%d error=%v", len(results), err)
		}
	}
}

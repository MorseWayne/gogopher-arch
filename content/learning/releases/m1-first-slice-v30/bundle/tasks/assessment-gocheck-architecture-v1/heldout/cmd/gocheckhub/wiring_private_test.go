package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompositionRootContract(t *testing.T) {
	handler, err := buildHandler()
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/checks", strings.NewReader(`{"target":"wired.example.com"}`))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusCreated || !strings.Contains(string(body), `"target":"wired.example.com"`) {
		t.Fatalf("wired response = status %d body %q", response.StatusCode, body)
	}

	duplicate := httptest.NewRecorder()
	handler.ServeHTTP(duplicate, httptest.NewRequest(http.MethodPost, "/checks", strings.NewReader(`{"target":"WIRED.EXAMPLE.COM"}`)))
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("wired duplicate status = %d", duplicate.Code)
	}
}

package httpapi_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocheckhub/internal/checks"
	"gocheckhub/internal/httpapi"
)

type creatorStub struct {
	input checks.NewCheck
	check checks.Check
	err   error
}

func (s *creatorStub) Create(_ context.Context, input checks.NewCheck) (checks.Check, error) {
	s.input = input
	return s.check, s.err
}

func TestTransportContract(t *testing.T) {
	if _, err := httpapi.NewHandler(nil); err == nil {
		t.Fatal("NewHandler accepted nil creator")
	}
	tests := []struct {
		name       string
		body       string
		creatorErr error
		created    checks.Check
		wantStatus int
		wantBody   string
	}{
		{name: "created", body: `{"target":"api.example.com"}`, created: checks.Check{ID: "check-7", Target: "api.example.com"}, wantStatus: 201, wantBody: `{"id":"check-7","target":"api.example.com"}` + "\n"},
		{name: "unknown field", body: `{"target":"api","extra":true}`, wantStatus: 400, wantBody: `{"error":{"code":"invalid_request","message":"request is invalid"}}` + "\n"},
		{name: "trailing value", body: `{"target":"api"} {}`, wantStatus: 400, wantBody: `{"error":{"code":"invalid_request","message":"request is invalid"}}` + "\n"},
		{name: "empty", body: `{"target":" "}`, creatorErr: checks.ErrInvalidTarget, wantStatus: 400, wantBody: `{"error":{"code":"invalid_request","message":"request is invalid"}}` + "\n"},
		{name: "conflict", body: `{"target":"api"}`, creatorErr: checks.ErrCheckExists, wantStatus: 409, wantBody: `{"error":{"code":"check_exists","message":"check already exists"}}` + "\n"},
		{name: "internal", body: `{"target":"api"}`, creatorErr: errors.New("secret database address"), wantStatus: 500, wantBody: `{"error":{"code":"internal_error","message":"internal server error"}}` + "\n"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			creator := &creatorStub{check: testCase.created, err: testCase.creatorErr}
			handler, err := httpapi.NewHandler(creator)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/checks", strings.NewReader(testCase.body))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != testCase.wantStatus || string(body) != testCase.wantBody {
				t.Fatalf("response = status %d body %q", response.StatusCode, body)
			}
			if got := response.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q", got)
			}
		})
	}
}

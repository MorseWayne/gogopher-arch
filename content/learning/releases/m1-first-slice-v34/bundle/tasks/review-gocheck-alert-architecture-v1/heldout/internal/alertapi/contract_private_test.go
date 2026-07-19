package alertapi_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocheckhub/internal/alertapi"
	"gocheckhub/internal/alerts"
)

type publisherStub struct {
	rule alerts.Rule
	err  error
}

func (s *publisherStub) Publish(_ context.Context, _ alerts.NewRule) (alerts.Rule, error) {
	return s.rule, s.err
}

func TestTransportContract(t *testing.T) {
	if _, err := alertapi.NewHandler(nil); err == nil {
		t.Fatal("NewHandler accepted nil publisher")
	}
	tests := []struct {
		name       string
		body       string
		result     alerts.Rule
		publishErr error
		wantStatus int
		wantBody   string
	}{
		{name: "created", body: `{"destination":"ops@example.com"}`, result: alerts.Rule{ID: "alert-7", Destination: "ops@example.com"}, wantStatus: 201, wantBody: `{"id":"alert-7","destination":"ops@example.com"}` + "\n"},
		{name: "unknown field", body: `{"destination":"ops@example.com","extra":true}`, wantStatus: 400, wantBody: `{"error":{"code":"invalid_request","message":"request is invalid"}}` + "\n"},
		{name: "trailing value", body: `{"destination":"ops@example.com"} {}`, wantStatus: 400, wantBody: `{"error":{"code":"invalid_request","message":"request is invalid"}}` + "\n"},
		{name: "invalid", body: `{"destination":" "}`, publishErr: alerts.ErrInvalidDestination, wantStatus: 400, wantBody: `{"error":{"code":"invalid_request","message":"request is invalid"}}` + "\n"},
		{name: "conflict", body: `{"destination":"ops@example.com"}`, publishErr: alerts.ErrRuleExists, wantStatus: 409, wantBody: `{"error":{"code":"alert_exists","message":"alert rule already exists"}}` + "\n"},
		{name: "internal", body: `{"destination":"ops@example.com"}`, publishErr: errors.New("secret storage path"), wantStatus: 500, wantBody: `{"error":{"code":"internal_error","message":"internal server error"}}` + "\n"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			publisher := &publisherStub{rule: testCase.result, err: testCase.publishErr}
			handler, err := alertapi.NewHandler(publisher)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(testCase.body))
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

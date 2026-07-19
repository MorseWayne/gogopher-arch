package quality

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type fixedClock struct{ value time.Time }

func (clock fixedClock) Now() time.Time { return clock.value }

type fixedIDs struct{ value string }

func (ids fixedIDs) NewID() string { return ids.value }

type recordingStore struct {
	saved []Check
	err   error
}

func (store *recordingStore) Save(_ context.Context, check Check) error {
	store.saved = append(store.saved, check)
	return store.err
}

func TestUnitBoundarySupportsDeterministicClockAndStoreFake(t *testing.T) {
	wantTime := time.Date(2026, time.July, 19, 0, 0, 0, 0, time.UTC)
	store := &recordingStore{}
	service, err := NewService(store, fixedClock{wantTime}, fixedIDs{"check-42"})
	if err != nil {
		t.Fatal(err)
	}
	check, err := service.Create(context.Background(), " api ", " https://example.test ")
	if err != nil || check.ID != "check-42" || !check.CreatedAt.Equal(wantTime) || len(store.saved) != 1 {
		t.Fatalf("check=%#v saved=%#v err=%v", check, store.saved, err)
	}
	if _, err := service.Create(context.Background(), "", "target"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid error=%v", err)
	}
	store.err = errors.New("write failed")
	if _, err := service.Create(context.Background(), "api", "target"); err == nil {
		t.Fatal("store error lost")
	}
}

type creatorFake struct {
	check  Check
	err    error
	called bool
}

func (fake *creatorFake) Create(context.Context, string, string) (Check, error) {
	fake.called = true
	return fake.check, fake.err
}

func TestHandlerBoundaryUsesServiceFakeAndStableProtocol(t *testing.T) {
	for _, test := range []struct {
		name, body string
		err        error
		want       int
		code       string
	}{{"created", `{"name":"api","target":"https://example.test"}`, nil, http.StatusCreated, ""}, {"bad json", `{`, nil, http.StatusBadRequest, "invalid_request"}, {"domain", `{"name":"","target":"x"}`, ErrInvalid, http.StatusBadRequest, "invalid_check"}, {"internal", `{"name":"api","target":"x"}`, errors.New("secret database error"), http.StatusInternalServerError, "internal_error"}} {
		t.Run(test.name, func(t *testing.T) {
			creator := &creatorFake{check: Check{ID: "check-1"}, err: test.err}
			handler, _ := NewHandler(creator)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/checks", bytes.NewBufferString(test.body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if test.code != "" {
				var body map[string]string
				_ = json.Unmarshal(response.Body.Bytes(), &body)
				if body["code"] != test.code {
					t.Fatalf("body=%v", body)
				}
			}
		})
	}
}

func TestFixtureSupportsLayeredScenarios(t *testing.T) {
	contents, err := os.ReadFile("internal/quality/testdata/checks.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]string
	if err := json.Unmarshal(contents, &rows); err != nil || len(rows) < 2 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	for _, row := range rows {
		if row["name"] == "" || row["target"] == "" {
			t.Fatalf("invalid fixture row=%v", row)
		}
	}
}

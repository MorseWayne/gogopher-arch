package alertquality

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

type hClock struct{ value time.Time }

func (clock hClock) Now() time.Time { return clock.value }

type hIDs struct{ value string }

func (ids hIDs) NewID() string { return ids.value }

type hStore struct {
	saved []Alert
	err   error
}

func (store *hStore) Save(_ context.Context, alert Alert) error {
	store.saved = append(store.saved, alert)
	return store.err
}
func TestUnitBoundarySupportsDeterministicClockAndStoreFake(t *testing.T) {
	store := &hStore{}
	service, _ := NewService(store, hClock{time.Unix(99, 0)}, hIDs{"alert-9"})
	alert, err := service.Create(context.Background(), " latency ", " https://hook.test ")
	if err != nil || alert.ID != "alert-9" || len(store.saved) != 1 {
		t.Fatalf("alert=%#v saved=%#v err=%v", alert, store.saved, err)
	}
	if _, err := service.Create(context.Background(), "", "x"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("error=%v", err)
	}
}

type hCreator struct {
	alert Alert
	err   error
}

func (fake hCreator) Create(context.Context, string, string) (Alert, error) {
	return fake.alert, fake.err
}
func TestHandlerBoundaryUsesServiceFakeAndStableProtocol(t *testing.T) {
	for _, test := range []struct {
		name, body string
		err        error
		want       int
	}{{"created", `{"name":"latency","destination":"https://hook.test"}`, nil, 201}, {"json", `{`, nil, 400}, {"domain", `{"name":"","destination":"x"}`, ErrInvalid, 400}, {"internal", `{"name":"x","destination":"y"}`, errors.New("secret"), 500}} {
		t.Run(test.name, func(t *testing.T) {
			handler, _ := NewHandler(hCreator{Alert{ID: "a"}, test.err})
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(test.body)))
			if response.Code != test.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
func TestFixtureSupportsLayeredScenarios(t *testing.T) {
	contents, err := os.ReadFile("internal/alertquality/testdata/alerts.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]string
	if err := json.Unmarshal(contents, &rows); err != nil || len(rows) < 2 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}

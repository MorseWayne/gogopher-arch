package alertquality

import (
	"context"
	"testing"
	"time"
)

type vClock struct{ value time.Time }

func (clock vClock) Now() time.Time { return clock.value }

type vIDs struct{}

func (vIDs) NewID() string { return "alert-1" }

type vStore struct{ saved Alert }

func (store *vStore) Save(_ context.Context, alert Alert) error { store.saved = alert; return nil }
func TestAlertSliceAcceptsDeterministicFakes(t *testing.T) {
	store := &vStore{}
	service, err := NewService(store, vClock{time.Unix(42, 0)}, vIDs{})
	if err != nil {
		t.Fatal(err)
	}
	alert, err := service.Create(context.Background(), " latency ", " https://hook.test ")
	if err != nil || alert.ID != "alert-1" || store.saved.Name != "latency" {
		t.Fatalf("alert=%#v saved=%#v err=%v", alert, store.saved, err)
	}
}

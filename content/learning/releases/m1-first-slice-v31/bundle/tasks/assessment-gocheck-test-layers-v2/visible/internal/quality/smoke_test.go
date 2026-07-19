package quality

import (
	"context"
	"testing"
	"time"
)

type smokeClock struct{ value time.Time }

func (clock smokeClock) Now() time.Time { return clock.value }

type smokeIDs struct{}

func (smokeIDs) NewID() string { return "check-1" }

type smokeStore struct{ saved Check }

func (store *smokeStore) Save(_ context.Context, check Check) error { store.saved = check; return nil }

func TestProductionSliceSupportsDeterministicFakes(t *testing.T) {
	store := &smokeStore{}
	wantTime := time.Unix(123, 0)
	service, err := NewService(store, smokeClock{wantTime}, smokeIDs{})
	if err != nil {
		t.Fatal(err)
	}
	check, err := service.Create(context.Background(), " homepage ", " https://example.test ")
	if err != nil || check.ID != "check-1" || !check.CreatedAt.Equal(wantTime) || store.saved.Name != "homepage" {
		t.Fatalf("check=%#v saved=%#v err=%v", check, store.saved, err)
	}
}

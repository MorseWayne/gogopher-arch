package alertcache

import (
	"context"
	"testing"
	"time"
)

type repositoryStub struct{}

func (repositoryStub) Find(context.Context, string) (Rule, error) {
	return Rule{ID: "a1", Destination: "ops", Version: 1}, nil
}
func (repositoryStub) Save(_ context.Context, r Rule) (Rule, error) { return r, nil }

type cacheStub struct{}

func (cacheStub) Get(context.Context, string) (Entry, bool, error)        { return Entry{}, false, nil }
func (cacheStub) Set(context.Context, string, Entry, time.Duration) error { return nil }
func (cacheStub) Delete(context.Context, string) error                    { return nil }
func TestColdReadLoadsRule(t *testing.T) {
	service, err := New(repositoryStub{}, cacheStub{}, Options{PositiveTTL: time.Minute, NegativeTTL: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	rule, err := service.Get(t.Context(), "a1")
	if err != nil || rule.Destination != "ops" {
		t.Fatalf("rule=%+v err=%v", rule, err)
	}
}

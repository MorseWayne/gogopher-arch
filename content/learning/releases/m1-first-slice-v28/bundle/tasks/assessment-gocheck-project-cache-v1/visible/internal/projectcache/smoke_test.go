package projectcache

import (
	"context"
	"testing"
	"time"
)

type sourceStub struct{}

func (sourceStub) Get(context.Context, string) (Project, error) {
	return Project{ID: "p1", Name: "gopher", Version: 1}, nil
}
func (sourceStub) Update(_ context.Context, p Project) (Project, error) { return p, nil }

type cacheStub struct{}

func (cacheStub) Get(context.Context, string) (CacheEntry, bool, error) {
	return CacheEntry{}, false, nil
}
func (cacheStub) Set(context.Context, string, CacheEntry, time.Duration) error { return nil }
func (cacheStub) Delete(context.Context, string) error                         { return nil }
func TestColdReadLoadsProject(t *testing.T) {
	service, err := New(sourceStub{}, cacheStub{}, Options{PositiveTTL: time.Minute, NegativeTTL: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	project, err := service.Get(t.Context(), "p1")
	if err != nil || project.Name != "gopher" {
		t.Fatalf("project=%+v err=%v", project, err)
	}
}

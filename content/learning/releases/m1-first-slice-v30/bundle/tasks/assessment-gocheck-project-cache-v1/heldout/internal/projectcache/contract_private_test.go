package projectcache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type sourceFake struct {
	mu          sync.Mutex
	project     Project
	getErr      error
	updateErr   error
	gets        int
	updates     int
	getStarted  chan struct{}
	releaseGet  chan struct{}
	updateOrder *[]string
}

func (source *sourceFake) Get(ctx context.Context, _ string) (Project, error) {
	source.mu.Lock()
	source.gets++
	started, release := source.getStarted, source.releaseGet
	project, err := source.project, source.getErr
	source.mu.Unlock()
	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if release != nil {
		select {
		case <-release:
		case <-ctx.Done():
			return Project{}, ctx.Err()
		}
	}
	return project, err
}

func (source *sourceFake) Update(_ context.Context, project Project) (Project, error) {
	source.mu.Lock()
	defer source.mu.Unlock()
	source.updates++
	if source.updateOrder != nil {
		*source.updateOrder = append(*source.updateOrder, "source")
	}
	if source.updateErr != nil {
		return Project{}, source.updateErr
	}
	source.project = project
	return project, nil
}

func (source *sourceFake) counts() (int, int) {
	source.mu.Lock()
	defer source.mu.Unlock()
	return source.gets, source.updates
}

type cacheFake struct {
	mu          sync.Mutex
	entry       CacheEntry
	hit         bool
	getErr      error
	setErr      error
	deleteErr   error
	sets        []cacheSet
	deletes     int
	deleteOrder *[]string
}

type cacheSet struct {
	entry CacheEntry
	ttl   time.Duration
}

func (cache *cacheFake) Get(context.Context, string) (CacheEntry, bool, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return cache.entry, cache.hit, cache.getErr
}

func (cache *cacheFake) Set(_ context.Context, _ string, entry CacheEntry, ttl time.Duration) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.sets = append(cache.sets, cacheSet{entry: entry, ttl: ttl})
	return cache.setErr
}

func (cache *cacheFake) Delete(context.Context, string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.deletes++
	if cache.deleteOrder != nil {
		*cache.deleteOrder = append(*cache.deleteOrder, "cache")
	}
	return cache.deleteErr
}

func cacheOptions() Options {
	return Options{PositiveTTL: time.Minute, NegativeTTL: 5 * time.Second}
}

func TestFreshPositiveAndNegativeHitsAvoidSource(t *testing.T) {
	tests := []struct {
		name    string
		entry   CacheEntry
		wantErr error
	}{
		{"positive", CacheEntry{Project: Project{ID: "p1", Name: "cached"}, Found: true}, nil},
		{"negative", CacheEntry{Found: false}, ErrNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := &sourceFake{project: Project{ID: "p1", Name: "truth"}}
			service, _ := New(source, &cacheFake{entry: test.entry, hit: true}, cacheOptions())
			project, err := service.Get(t.Context(), "p1")
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("project=%+v err=%v", project, err)
			}
			if gets, _ := source.counts(); gets != 0 {
				t.Fatalf("source gets=%d", gets)
			}
		})
	}
}

func TestMissLoadsTruthWithPositiveAndNegativeTTL(t *testing.T) {
	tests := []struct {
		name    string
		project Project
		source  error
		found   bool
		ttl     time.Duration
	}{
		{"positive", Project{ID: "p1", Name: "truth"}, nil, true, time.Minute},
		{"negative", Project{}, ErrNotFound, false, 5 * time.Second},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := &cacheFake{}
			service, _ := New(&sourceFake{project: test.project, getErr: test.source}, cache, cacheOptions())
			_, err := service.Get(t.Context(), "p1")
			if !errors.Is(err, test.source) {
				t.Fatalf("err=%v", err)
			}
			cache.mu.Lock()
			defer cache.mu.Unlock()
			if len(cache.sets) != 1 || cache.sets[0].entry.Found != test.found || cache.sets[0].ttl != test.ttl {
				t.Fatalf("sets=%+v", cache.sets)
			}
		})
	}
}

func TestCacheFailureDegradesToSource(t *testing.T) {
	cacheDown := errors.New("cache down")
	cache := &cacheFake{getErr: cacheDown, setErr: cacheDown}
	source := &sourceFake{project: Project{ID: "p1", Name: "truth"}}
	service, _ := New(source, cache, cacheOptions())
	project, err := service.Get(t.Context(), "p1")
	if err != nil || project.Name != "truth" {
		t.Fatalf("project=%+v err=%v", project, err)
	}
	if gets, _ := source.counts(); gets != 1 {
		t.Fatalf("source gets=%d", gets)
	}
}

func TestConcurrentMissCoalescesAndWaiterCancels(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	source := &sourceFake{project: Project{ID: "p1", Name: "truth"}, getStarted: started, releaseGet: release}
	service, _ := New(source, &cacheFake{}, cacheOptions())
	leader := make(chan error, 1)
	go func() {
		_, err := service.Get(context.Background(), "p1")
		leader <- err
	}()
	<-started
	waiterContext, cancel := context.WithCancel(context.Background())
	waiter := make(chan error, 1)
	go func() {
		_, err := service.Get(waiterContext, "p1")
		waiter <- err
	}()
	cancel()
	select {
	case err := <-waiter:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("waiter err=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter ignored context")
	}
	close(release)
	if err := <-leader; err != nil {
		t.Fatal(err)
	}
	if gets, _ := source.counts(); gets != 1 {
		t.Fatalf("source gets=%d", gets)
	}
}

func TestUpdateWritesTruthBeforeInvalidation(t *testing.T) {
	order := []string{}
	source := &sourceFake{updateOrder: &order}
	cache := &cacheFake{deleteOrder: &order}
	service, _ := New(source, cache, cacheOptions())
	updated, err := service.Update(t.Context(), Project{ID: "p1", Name: "new"})
	if err != nil || updated.Name != "new" {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if len(order) != 2 || order[0] != "source" || order[1] != "cache" {
		t.Fatalf("order=%v", order)
	}
}

func TestUpdateFailureAndInvalidationFailureRemainDistinct(t *testing.T) {
	truthFailure := errors.New("truth failed")
	cacheFailure := errors.New("delete failed")
	t.Run("truth failure does not invalidate", func(t *testing.T) {
		source := &sourceFake{updateErr: truthFailure}
		cache := &cacheFake{}
		service, _ := New(source, cache, cacheOptions())
		if _, err := service.Update(t.Context(), Project{ID: "p1"}); !errors.Is(err, truthFailure) {
			t.Fatalf("err=%v", err)
		}
		if cache.deletes != 0 {
			t.Fatalf("deletes=%d", cache.deletes)
		}
	})
	t.Run("invalidation failure exposes committed truth", func(t *testing.T) {
		source := &sourceFake{}
		cache := &cacheFake{deleteErr: cacheFailure}
		service, _ := New(source, cache, cacheOptions())
		updated, err := service.Update(t.Context(), Project{ID: "p1", Name: "committed"})
		if updated.Name != "committed" || !errors.Is(err, cacheFailure) {
			t.Fatalf("updated=%+v err=%v", updated, err)
		}
	})
}

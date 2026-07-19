package alertcache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type repositoryFake struct {
	mu        sync.Mutex
	rule      Rule
	findErr   error
	saveErr   error
	finds     int
	started   chan struct{}
	release   chan struct{}
	saveOrder *[]string
}

func (repository *repositoryFake) Find(ctx context.Context, _ string) (Rule, error) {
	repository.mu.Lock()
	repository.finds++
	started, release := repository.started, repository.release
	rule, err := repository.rule, repository.findErr
	repository.mu.Unlock()
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
			return Rule{}, ctx.Err()
		}
	}
	return rule, err
}

func (repository *repositoryFake) Save(_ context.Context, rule Rule) (Rule, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.saveOrder != nil {
		*repository.saveOrder = append(*repository.saveOrder, "repository")
	}
	if repository.saveErr != nil {
		return Rule{}, repository.saveErr
	}
	repository.rule = rule
	return rule, nil
}

func (repository *repositoryFake) findCount() int {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	return repository.finds
}

type alertCacheFake struct {
	mu          sync.Mutex
	entry       Entry
	hit         bool
	getErr      error
	setErr      error
	deleteErr   error
	sets        []alertSet
	deletes     int
	deleteOrder *[]string
}

type alertSet struct {
	entry Entry
	ttl   time.Duration
}

func (cache *alertCacheFake) Get(context.Context, string) (Entry, bool, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return cache.entry, cache.hit, cache.getErr
}
func (cache *alertCacheFake) Set(_ context.Context, _ string, entry Entry, ttl time.Duration) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.sets = append(cache.sets, alertSet{entry, ttl})
	return cache.setErr
}
func (cache *alertCacheFake) Delete(context.Context, string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.deletes++
	if cache.deleteOrder != nil {
		*cache.deleteOrder = append(*cache.deleteOrder, "cache")
	}
	return cache.deleteErr
}

func alertOptions() Options { return Options{PositiveTTL: time.Minute, NegativeTTL: 5 * time.Second} }

func TestFreshAlertHitsAvoidRepository(t *testing.T) {
	for _, test := range []struct {
		name  string
		entry Entry
		want  error
	}{
		{"positive", Entry{Rule: Rule{ID: "a1", Destination: "cached"}, Found: true}, nil},
		{"negative", Entry{Found: false}, ErrNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &repositoryFake{rule: Rule{ID: "a1", Destination: "truth"}}
			service, _ := New(repository, &alertCacheFake{entry: test.entry, hit: true}, alertOptions())
			_, err := service.Get(t.Context(), "a1")
			if !errors.Is(err, test.want) || repository.findCount() != 0 {
				t.Fatalf("err=%v finds=%d", err, repository.findCount())
			}
		})
	}
}

func TestAlertMissCachesPositiveAndNegativeTTL(t *testing.T) {
	for _, test := range []struct {
		name   string
		rule   Rule
		source error
		found  bool
		ttl    time.Duration
	}{
		{"positive", Rule{ID: "a1", Destination: "ops"}, nil, true, time.Minute},
		{"negative", Rule{}, ErrNotFound, false, 5 * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			cache := &alertCacheFake{}
			service, _ := New(&repositoryFake{rule: test.rule, findErr: test.source}, cache, alertOptions())
			_, err := service.Get(t.Context(), "a1")
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

func TestAlertCacheOutageDegradesToRepository(t *testing.T) {
	down := errors.New("cache down")
	repository := &repositoryFake{rule: Rule{ID: "a1", Destination: "truth"}}
	service, _ := New(repository, &alertCacheFake{getErr: down, setErr: down}, alertOptions())
	rule, err := service.Get(t.Context(), "a1")
	if err != nil || rule.Destination != "truth" || repository.findCount() != 1 {
		t.Fatalf("rule=%+v err=%v finds=%d", rule, err, repository.findCount())
	}
}

func TestAlertConcurrentMissCoalescesAndWaiterCancels(t *testing.T) {
	started, release := make(chan struct{}, 1), make(chan struct{})
	repository := &repositoryFake{rule: Rule{ID: "a1"}, started: started, release: release}
	service, _ := New(repository, &alertCacheFake{}, alertOptions())
	leader := make(chan error, 1)
	go func() { _, err := service.Get(context.Background(), "a1"); leader <- err }()
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	waiter := make(chan error, 1)
	go func() { _, err := service.Get(ctx, "a1"); waiter <- err }()
	cancel()
	select {
	case err := <-waiter:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter ignored context")
	}
	close(release)
	if err := <-leader; err != nil {
		t.Fatal(err)
	}
	if repository.findCount() != 1 {
		t.Fatalf("finds=%d", repository.findCount())
	}
}

func TestSaveWritesRepositoryBeforeInvalidation(t *testing.T) {
	order := []string{}
	repository := &repositoryFake{saveOrder: &order}
	cache := &alertCacheFake{deleteOrder: &order}
	service, _ := New(repository, cache, alertOptions())
	saved, err := service.Save(t.Context(), Rule{ID: "a1", Destination: "new"})
	if err != nil || saved.Destination != "new" || len(order) != 2 || order[0] != "repository" || order[1] != "cache" {
		t.Fatalf("saved=%+v order=%v err=%v", saved, order, err)
	}
}

func TestSaveFailureBoundariesRemainDistinct(t *testing.T) {
	repositoryFailure, cacheFailure := errors.New("repository failed"), errors.New("delete failed")
	t.Run("repository failure", func(t *testing.T) {
		repository := &repositoryFake{saveErr: repositoryFailure}
		cache := &alertCacheFake{}
		service, _ := New(repository, cache, alertOptions())
		if _, err := service.Save(t.Context(), Rule{ID: "a1"}); !errors.Is(err, repositoryFailure) {
			t.Fatalf("err=%v", err)
		}
		if cache.deletes != 0 {
			t.Fatalf("deletes=%d", cache.deletes)
		}
	})
	t.Run("invalidation failure", func(t *testing.T) {
		repository := &repositoryFake{}
		cache := &alertCacheFake{deleteErr: cacheFailure}
		service, _ := New(repository, cache, alertOptions())
		saved, err := service.Save(t.Context(), Rule{ID: "a1", Destination: "committed"})
		if saved.Destination != "committed" || !errors.Is(err, cacheFailure) {
			t.Fatalf("saved=%+v err=%v", saved, err)
		}
	})
}

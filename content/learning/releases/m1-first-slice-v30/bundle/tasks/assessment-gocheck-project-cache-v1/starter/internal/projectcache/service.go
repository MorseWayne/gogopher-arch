package projectcache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("project not found")

type Project struct {
	ID, Name string
	Version  int
}
type CacheEntry struct {
	Project Project
	Found   bool
}
type Source interface {
	Get(context.Context, string) (Project, error)
	Update(context.Context, Project) (Project, error)
}
type Cache interface {
	Get(context.Context, string) (CacheEntry, bool, error)
	Set(context.Context, string, CacheEntry, time.Duration) error
	Delete(context.Context, string) error
}
type Options struct{ PositiveTTL, NegativeTTL time.Duration }
type flight struct {
	done    chan struct{}
	project Project
	err     error
}
type Service struct {
	source  Source
	cache   Cache
	options Options
	mu      sync.Mutex
	flights map[string]*flight
}

func New(source Source, cache Cache, options Options) (*Service, error) {
	// TODO: Please implement
	return nil, errors.New("TODO")
}
func (service *Service) Get(ctx context.Context, id string) (Project, error) {
	// TODO: cache-aside lookup with degradation and miss coalescing.
	return Project{}, errors.New("TODO")
}
func (service *Service) Update(ctx context.Context, project Project) (Project, error) {
	// TODO: write source of truth before invalidating cache.
	return Project{}, errors.New("TODO")
}

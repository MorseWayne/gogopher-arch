package alertcache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("alert rule not found")

type Rule struct {
	ID, Destination string
	Version         int
}
type Entry struct {
	Rule  Rule
	Found bool
}
type Repository interface {
	Find(context.Context, string) (Rule, error)
	Save(context.Context, Rule) (Rule, error)
}
type Cache interface {
	Get(context.Context, string) (Entry, bool, error)
	Set(context.Context, string, Entry, time.Duration) error
	Delete(context.Context, string) error
}
type Options struct{ PositiveTTL, NegativeTTL time.Duration }
type flight struct {
	done chan struct{}
	rule Rule
	err  error
}
type Service struct {
	repository Repository
	cache      Cache
	options    Options
	mu         sync.Mutex
	flights    map[string]*flight
}

func New(repository Repository, cache Cache, options Options) (*Service, error) {
	return nil, errors.New("TODO")
}
func (service *Service) Get(ctx context.Context, id string) (Rule, error) {
	return Rule{}, errors.New("TODO")
}
func (service *Service) Save(ctx context.Context, rule Rule) (Rule, error) {
	return Rule{}, errors.New("TODO")
}

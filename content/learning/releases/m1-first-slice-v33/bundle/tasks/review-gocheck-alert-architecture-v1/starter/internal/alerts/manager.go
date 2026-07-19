package alerts

import (
	"context"
	"errors"
)

var (
	ErrInvalidDestination = errors.New("invalid destination")
	ErrRuleExists         = errors.New("alert rule already exists")
)

type NewRule struct {
	Destination string
}

type Rule struct {
	ID          string
	Destination string
}

type Store interface {
	Save(context.Context, Rule) error
}

type Manager struct {
	store  Store
	nextID func() string
}

func NewManager(store Store, nextID func() string) (*Manager, error) {
	return nil, errors.New("TODO: validate and store use case dependencies")
}

func (m *Manager) Publish(ctx context.Context, input NewRule) (Rule, error) {
	return Rule{}, errors.New("TODO: apply rules and save the alert")
}

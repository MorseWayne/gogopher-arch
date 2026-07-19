package checks

import (
	"context"
	"errors"
)

var (
	ErrInvalidTarget = errors.New("invalid target")
	ErrCheckExists   = errors.New("check already exists")
)

type NewCheck struct {
	Target string
}

type Check struct {
	ID     string
	Target string
}

type Repository interface {
	Create(context.Context, Check) error
}

type Service struct {
	repository Repository
	nextID     func() string
}

func NewService(repository Repository, nextID func() string) (*Service, error) {
	return nil, errors.New("TODO: validate and store use case dependencies")
}

func (s *Service) Create(ctx context.Context, input NewCheck) (Check, error) {
	return Check{}, errors.New("TODO: apply rules and persist the check")
}

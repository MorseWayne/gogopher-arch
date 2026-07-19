package alertboard

import (
	"context"
	"errors"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("not found")

type Alert struct { ID, TenantID, Message string; Acknowledged bool }
type Delivery struct { ID, TenantID, Target string }
type Store interface {
	Alert(context.Context, string, string) (Alert, error)
	Acknowledge(context.Context, string, string) error
	Next(context.Context) (Delivery, error)
	Complete(context.Context, string, error) error
}
type Cache interface {
	Get(context.Context, string) (Alert, bool, error)
	Set(context.Context, string, Alert, time.Duration) error
	Delete(context.Context, string) error
}

type Service struct{}

func NewService(Store, Cache, map[string]string) (*Service, error) { return &Service{}, nil }
func (service *Service) Handler() http.Handler { return http.NewServeMux() }
func (service *Service) Run(context.Context, int, func(context.Context, Delivery) error) error { return nil }

package alertaccess

import (
	"context"
	"errors"
	"net/http"
)

var ErrNotFound = errors.New("alert rule not found")

type Credential struct {
	Subject string
	Token   string
}

type Rule struct {
	ID      string
	OwnerID string
}

type RuleStore interface {
	FindRule(context.Context, string) (Rule, error)
	DeleteRule(context.Context, string) error
}

type AuditLogger interface {
	Denied(context.Context, string)
}

type Handler struct{}

func New(credentials []Credential, store RuleStore, logger AuditLogger) (*Handler, error) {
	return nil, errors.New("TODO: validate and hash credentials")
}

func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	http.Error(response, "TODO", http.StatusNotImplemented)
}

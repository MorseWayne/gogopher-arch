package projectapi

import (
	"context"
	"errors"
	"net/http"
)

var ErrNotFound = errors.New("project not found")

type Credential struct {
	Subject string
	Token   string
}

type Project struct {
	ID      string `json:"id"`
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
}

type ProjectStore interface {
	FindProject(context.Context, string) (Project, error)
}

type AuditLogger interface {
	Denied(context.Context, string)
}

type Handler struct{}

func New(credentials []Credential, store ProjectStore, logger AuditLogger) (*Handler, error) {
	return nil, errors.New("TODO: validate and hash credentials")
}

func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	http.Error(response, "TODO", http.StatusNotImplemented)
}

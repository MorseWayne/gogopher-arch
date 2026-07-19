package httpapi

import (
	"context"
	"errors"
	"net/http"

	"gocheckhub/internal/checks"
)

type Creator interface {
	Create(context.Context, checks.NewCheck) (checks.Check, error)
}

func NewHandler(creator Creator) (http.Handler, error) {
	return nil, errors.New("TODO: build the checks transport")
}

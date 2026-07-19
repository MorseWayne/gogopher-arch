package alertapi

import (
	"context"
	"errors"
	"net/http"

	"gocheckhub/internal/alerts"
)

type Publisher interface {
	Publish(context.Context, alerts.NewRule) (alerts.Rule, error)
}

func NewHandler(publisher Publisher) (http.Handler, error) {
	return nil, errors.New("TODO: build the alerts transport")
}

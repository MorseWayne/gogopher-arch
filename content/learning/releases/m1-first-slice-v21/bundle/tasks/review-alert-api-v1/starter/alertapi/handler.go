package alertapi

import (
	"context"
	"errors"
	"net/http"
)

var ErrRuleExists = errors.New("rule already exists")
type NewRule struct { Destination string; Threshold int }
type Rule struct { ID, Destination string; Threshold int }
type Creator interface { Create(context.Context, NewRule) (Rule, error) }

func NewHandler(_ Creator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rules", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "TODO", http.StatusNotImplemented) })
	return mux
}

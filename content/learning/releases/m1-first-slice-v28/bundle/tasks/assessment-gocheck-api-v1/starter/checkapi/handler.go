package checkapi

import (
	"context"
	"errors"
	"net/http"
)

var ErrCheckExists = errors.New("check already exists")

type NewCheck struct { Target string; TimeoutMS int }
type Check struct { ID, Target string; TimeoutMS int }
type Creator interface { Create(context.Context, NewCheck) (Check, error) }

func NewHandler(_ Creator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /checks", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "TODO", http.StatusNotImplemented) })
	return mux
}

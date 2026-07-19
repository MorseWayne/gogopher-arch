package quality

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid check")

type Check struct {
	ID, Name, Target string
	CreatedAt        time.Time
}
type Clock interface{ Now() time.Time }
type IDSource interface{ NewID() string }
type Store interface {
	Save(context.Context, Check) error
}

type Service struct {
	store Store
	clock Clock
	ids   IDSource
}

func NewService(store Store, clock Clock, ids IDSource) (*Service, error) {
	if store == nil || clock == nil || ids == nil {
		return nil, errors.New("invalid service dependencies")
	}
	return &Service{store: store, clock: clock, ids: ids}, nil
}

func (service *Service) Create(ctx context.Context, name, target string) (Check, error) {
	name, target = strings.TrimSpace(name), strings.TrimSpace(target)
	if name == "" || target == "" {
		return Check{}, ErrInvalid
	}
	check := Check{ID: service.ids.NewID(), Name: name, Target: target, CreatedAt: service.clock.Now()}
	if check.ID == "" {
		return Check{}, errors.New("empty generated id")
	}
	if err := service.store.Save(ctx, check); err != nil {
		return Check{}, err
	}
	return check, nil
}

type Creator interface {
	Create(context.Context, string, string) (Check, error)
}
type Handler struct{ creator Creator }

func NewHandler(creator Creator) (*Handler, error) {
	if creator == nil {
		return nil, errors.New("nil creator")
	}
	return &Handler{creator: creator}, nil
}

func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Name   string `json:"name"`
		Target string `json:"target"`
	}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request")
		return
	}
	check, err := handler.creator.Create(request.Context(), input.Name, input.Target)
	if errors.Is(err, ErrInvalid) {
		writeError(response, http.StatusBadRequest, "invalid_check")
		return
	}
	if err != nil {
		writeError(response, http.StatusInternalServerError, "internal_error")
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(response).Encode(check)
}

func writeError(response http.ResponseWriter, status int, code string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]string{"code": code})
}

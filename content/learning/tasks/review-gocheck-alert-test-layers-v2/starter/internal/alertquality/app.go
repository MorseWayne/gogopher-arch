package alertquality

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid alert")

type Alert struct {
	ID, Name, Destination string
	CreatedAt             time.Time
}
type Clock interface{ Now() time.Time }
type IDSource interface{ NewID() string }
type Store interface {
	Save(context.Context, Alert) error
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
	return &Service{store, clock, ids}, nil
}
func (service *Service) Create(ctx context.Context, name, destination string) (Alert, error) {
	name, destination = strings.TrimSpace(name), strings.TrimSpace(destination)
	if name == "" || destination == "" {
		return Alert{}, ErrInvalid
	}
	alert := Alert{ID: service.ids.NewID(), Name: name, Destination: destination, CreatedAt: service.clock.Now()}
	if alert.ID == "" {
		return Alert{}, errors.New("empty generated id")
	}
	if err := service.store.Save(ctx, alert); err != nil {
		return Alert{}, err
	}
	return alert, nil
}

type Creator interface {
	Create(context.Context, string, string) (Alert, error)
}
type Handler struct{ creator Creator }

func NewHandler(creator Creator) (*Handler, error) {
	if creator == nil {
		return nil, errors.New("nil creator")
	}
	return &Handler{creator}, nil
}
func (handler *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Name        string `json:"name"`
		Destination string `json:"destination"`
	}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		alertError(response, http.StatusBadRequest, "invalid_request")
		return
	}
	alert, err := handler.creator.Create(request.Context(), input.Name, input.Destination)
	if errors.Is(err, ErrInvalid) {
		alertError(response, http.StatusBadRequest, "invalid_alert")
		return
	}
	if err != nil {
		alertError(response, http.StatusInternalServerError, "internal_error")
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(response).Encode(alert)
}
func alertError(response http.ResponseWriter, status int, code string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]string{"code": code})
}

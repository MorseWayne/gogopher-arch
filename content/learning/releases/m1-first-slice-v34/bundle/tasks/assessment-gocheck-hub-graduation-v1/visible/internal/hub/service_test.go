package hub

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type visibleStore struct {
	mu       sync.Mutex
	projects map[string]Project
	lookups  int
	readyErr error
}

func (store *visibleStore) CreateProject(_ context.Context, project Project) (Project, error) {
	store.mu.Lock(); defer store.mu.Unlock()
	if store.projects == nil { store.projects = map[string]Project{} }
	store.projects[project.TenantID+"/"+project.ID] = project
	return project, nil
}
func (store *visibleStore) Project(_ context.Context, tenantID, id string) (Project, error) {
	store.mu.Lock(); defer store.mu.Unlock(); store.lookups++
	project, ok := store.projects[tenantID+"/"+id]
	if !ok { return Project{}, ErrNotFound }
	return project, nil
}
func (store *visibleStore) Ready(context.Context) error { return store.readyErr }
func (store *visibleStore) Claim(ctx context.Context) (Job, error) { <-ctx.Done(); return Job{}, ctx.Err() }
func (store *visibleStore) Complete(context.Context, string, error) error { return nil }

func newVisibleService(t *testing.T, store Store) *Service {
	t.Helper()
	service, err := NewService(store, nil, map[string]string{"tenant-a": "secret-a", "tenant-b": "secret-b"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil { t.Fatal(err) }
	return service
}

func TestPublicServiceContract(t *testing.T) {
	store := &visibleStore{projects: map[string]Project{"tenant-a/p-1": {ID: "p-1", TenantID: "tenant-a", Name: "docs"}}}
	handler := newVisibleService(t, store).Handler()
	t.Run("liveness", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/livez", nil)
		response := httptest.NewRecorder(); handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Header().Get("X-Request-ID") == "" { t.Fatalf("response = %#v", response.Result()) }
	})
	t.Run("auth before lookup", func(t *testing.T) {
		before := store.lookups
		response := httptest.NewRecorder(); handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/projects/p-1", nil))
		if response.Code != http.StatusUnauthorized || store.lookups != before { t.Fatalf("status=%d lookups=%d", response.Code, store.lookups-before) }
	})
	t.Run("strict create", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/v1/projects", strings.NewReader(`{"name":"api","unknown":true}`)); request.Header.Set("X-API-Key", "secret-a")
		response := httptest.NewRecorder(); handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s", response.Code, response.Body.String()) }
	})
}

var _ = errors.Is
var _ = time.Second

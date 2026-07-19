package hub

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type raceStore struct{}
func (raceStore) CreateProject(context.Context, Project) (Project, error) { return Project{}, nil }
func (raceStore) Project(context.Context, string, string) (Project, error) { return Project{}, ErrNotFound }
func (raceStore) Ready(context.Context) error { return nil }
func (raceStore) Claim(ctx context.Context) (Job, error) { <-ctx.Done(); return Job{}, ctx.Err() }
func (raceStore) Complete(context.Context, string, error) error { return nil }

func TestHandlersAreRaceClean(t *testing.T) {
	store := raceStore{}
	service, err := NewService(store, nil, map[string]string{"tenant-a":"secret"}, slog.New(slog.NewTextHandler(io.Discard,nil)))
	if err != nil { t.Fatal(err) }
	handler := service.Handler()
	var wait sync.WaitGroup
	for index:=0; index<24; index++ { wait.Add(1); go func(){ defer wait.Done(); response:=httptest.NewRecorder(); handler.ServeHTTP(response,httptest.NewRequest(http.MethodGet,"/livez",nil)); if response.Code!=200 { t.Errorf("status=%d",response.Code) } }() }
	wait.Wait()
}

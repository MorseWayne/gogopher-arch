package attempt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestPostgresRepositorySerializesConcurrentWorkspaceSaves(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := platformdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := fmt.Sprintf("attempt_test_%d", time.Now().UTC().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`)
	migrator, err := platformdb.NewMigrator(db, os.DirFS("../../../db/migrations"), platformdb.MigratorOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}
	contentDir, _ := filepath.Abs("../../../content/learning")
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	history, _ := definition.NewReleaseStore(db, definition.ReleaseStoreOptions{Schema: schema})
	if err := history.Register(ctx, registry); err != nil {
		t.Fatal(err)
	}
	learnerID := "00000000-0000-4000-8000-000000000101"
	if _, err := db.ExecContext(ctx, `INSERT INTO "`+schema+`".learners (id) VALUES ($1)`, learnerID); err != nil {
		t.Fatal(err)
	}
	repository, _ := NewPostgresRepository(db, RepositoryOptions{Schema: schema})
	service, err := NewService(repository, registry, ServiceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Create(ctx, CreateInput{LearnerID: learnerID, ActivityID: "assessment-check-config", ActivityVersion: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !created.Created {
		t.Fatalf("first Create() = %#v, want Created", created)
	}
	resumed, err := service.Create(ctx, CreateInput{LearnerID: learnerID, ActivityID: "assessment-check-config", ActivityVersion: 3})
	if err != nil || resumed.Created || resumed.ID != created.ID {
		t.Fatalf("second Create() = %#v, %v, want resumed %s", resumed, err, created.ID)
	}
	if _, err := service.Get(ctx, "00000000-0000-4000-8000-000000000999", created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner Get() error = %v", err)
	}

	files := cloneWorkspace(created.Workspace)
	files["internal/config/config.go"] += "\n// concurrent change\n"
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := service.Save(ctx, SaveInput{LearnerID: learnerID, AttemptID: created.ID, BaseRevision: 0, Files: files})
			results <- err
		}()
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		var conflict *RevisionConflict
		if errors.As(err, &conflict) {
			conflicts++
			continue
		}
		t.Fatalf("concurrent Save() error = %v", err)
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	final, err := service.Get(ctx, learnerID, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if final.WorkspaceRevision != 1 || final.WorkspaceHash != WorkspaceHash(files) {
		t.Fatalf("final attempt = %#v", final)
	}
}

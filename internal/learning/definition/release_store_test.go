package definition

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestReleaseStoreRegistersIdempotentlyAndRejectsVersionDrift(t *testing.T) {
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

	schema := fmt.Sprintf("definition_test_%d", time.Now().UTC().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA `+quotePostgresIdentifier(schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	defer db.ExecContext(context.Background(), `DROP SCHEMA `+quotePostgresIdentifier(schema)+` CASCADE`)
	migrator, err := platformdb.NewMigrator(db, os.DirFS("../../../db/migrations"), platformdb.MigratorOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}

	registry, err := LoadRegistry(RegistryOptions{ContentDir: repositoryContentDir(t)})
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewReleaseStore(db, ReleaseStoreOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Register(ctx, registry); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := store.Register(ctx, registry); err != nil {
		t.Fatalf("idempotent Register() error = %v", err)
	}
	current, err := store.CurrentReleaseID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if current != testReleaseID {
		t.Fatalf("CurrentReleaseID() = %q, want %q", current, testReleaseID)
	}
	var definitionCount int
	if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT count(*) FROM %s.definition_versions`, quotePostgresIdentifier(schema))).Scan(&definitionCount); err != nil {
		t.Fatal(err)
	}
	if definitionCount != 16 {
		t.Fatalf("definition_versions count = %d, want 16", definitionCount)
	}

	ref := DefinitionRef{ReleaseID: testReleaseID, Kind: KindCapability, ID: "M1-01", Version: 1}
	drifted := registry.definitions[ref]
	drifted.ContentHash = strings.Repeat("f", 64)
	registry.definitions[ref] = drifted
	if err := store.Register(ctx, registry); err == nil || !strings.Contains(err.Error(), "immutable database history") {
		t.Fatalf("Register(drifted) error = %v, want immutable history conflict", err)
	}

	learnerID := "00000000-0000-4000-8000-000000000001"
	attemptID := "00000000-0000-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO %s.learners (id) VALUES ($1)`, quotePostgresIdentifier(schema)), learnerID); err != nil {
		t.Fatal(err)
	}
	hash := strings.Repeat("a", 64)
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s.learning_attempts (
			id, learner_id, release_id, activity_id, activity_version, activity_hash,
			task_id, task_version, task_hash, capability_refs, mode, status, workspace, workspace_hash
		) VALUES ($1, $2, $3, 'assessment-check-config', 1, $4,
			'assessment-check-config-v1', 1, $4, '[]'::jsonb, 'assessment', 'active', '{}'::jsonb, $4)`, quotePostgresIdentifier(schema)), attemptID, learnerID, testReleaseID, hash); err != nil {
		t.Fatal(err)
	}
	referenced, err := store.ReferencedReleaseIDs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(referenced) != 1 || referenced[0] != testReleaseID {
		t.Fatalf("ReferencedReleaseIDs() = %#v, want [%s]", referenced, testReleaseID)
	}
}

func TestNewReleaseStoreValidatesDependencies(t *testing.T) {
	if _, err := NewReleaseStore(nil, ReleaseStoreOptions{}); err == nil {
		t.Fatal("NewReleaseStore(nil) error = nil")
	}
}

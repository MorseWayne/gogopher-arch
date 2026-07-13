package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestPostgresRepositoryPersistsOnlyTokenHashAndExpiresSession(t *testing.T) {
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

	schema := fmt.Sprintf("session_test_%d", time.Now().UTC().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA `+quoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(context.Background(), `DROP SCHEMA `+quoteIdentifier(schema)+` CASCADE`)
	migrator, err := platformdb.NewMigrator(db, os.DirFS("../../../db/migrations"), platformdb.MigratorOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}
	repository, err := NewPostgresRepository(db, RepositoryOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	service, err := NewService(repository, ServiceOptions{TTL: 24 * time.Hour, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}

	created, err := service.Establish(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	var storedHash string
	if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT token_hash FROM %s.learner_sessions WHERE id = $1`, quoteIdentifier(schema)), created.Session.ID).Scan(&storedHash); err != nil {
		t.Fatal(err)
	}
	if storedHash != TokenHash(created.Token) || storedHash == created.Token {
		t.Fatalf("stored token = %q, want only SHA-256 hash", storedHash)
	}
	reused, err := service.Establish(ctx, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if reused.Created || reused.Session.LearnerID != created.Session.LearnerID {
		t.Fatalf("reused session = %#v", reused)
	}
	if _, err := service.Authenticate(ctx, "forged-token"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Authenticate(forged) error = %v", err)
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.learner_sessions
		SET created_at = $2, last_used_at = $2, expires_at = $3
		WHERE id = $1`, quoteIdentifier(schema)), created.Session.ID, now.Add(-2*time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(ctx, created.Token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Authenticate(expired) error = %v", err)
	}
	replacement, err := service.Establish(ctx, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if !replacement.Created || replacement.Session.LearnerID == created.Session.LearnerID {
		t.Fatalf("replacement session = %#v", replacement)
	}
}

func TestNewPostgresRepositoryValidatesDependencies(t *testing.T) {
	if _, err := NewPostgresRepository(nil, RepositoryOptions{}); err == nil {
		t.Fatal("NewPostgresRepository(nil) error = nil")
	}
}

package database

import (
	"context"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestDiscoverMigrationsSortsAndChecksumsFiles(t *testing.T) {
	source := fstest.MapFS{
		"000002_add_attempts.up.sql":   {Data: []byte("CREATE TABLE attempts (id uuid PRIMARY KEY);\n")},
		"000001_add_learners.up.sql":   {Data: []byte("CREATE TABLE learners (id uuid PRIMARY KEY);\n")},
		"000001_add_learners.down.sql": {Data: []byte("DROP TABLE learners;\n")},
	}

	migrations, err := DiscoverMigrations(source)
	if err != nil {
		t.Fatalf("DiscoverMigrations() error = %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("len(migrations) = %d, want 2", len(migrations))
	}
	if migrations[0].Version != 1 || migrations[0].Name != "add_learners" {
		t.Fatalf("first migration = %#v, want version 1 add_learners", migrations[0])
	}
	if migrations[1].Version != 2 || migrations[1].Name != "add_attempts" {
		t.Fatalf("second migration = %#v, want version 2 add_attempts", migrations[1])
	}
	if migrations[0].Checksum != "b8b2285d1423da2c8c8bae6c5ab40db8337367322d8ec0f9ad4046822053bc96" {
		t.Fatalf("checksum = %q, want stable SHA-256", migrations[0].Checksum)
	}
}

func TestDiscoverMigrationsRejectsInvalidFiles(t *testing.T) {
	tests := map[string]fstest.MapFS{
		"invalid filename": {
			"1_users.up.sql": {Data: []byte("SELECT 1;")},
		},
		"duplicate version": {
			"000001_users.up.sql":    {Data: []byte("SELECT 1;")},
			"000001_sessions.up.sql": {Data: []byte("SELECT 2;")},
		},
		"empty migration": {
			"000001_users.up.sql": {Data: []byte("  \n")},
		},
	}

	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := DiscoverMigrations(source); err == nil {
				t.Fatal("DiscoverMigrations() error = nil, want validation error")
			}
		})
	}
}

func TestValidateAppliedRequiresAnExactPrefix(t *testing.T) {
	migrations := []Migration{
		{Version: 1, Name: "users", Checksum: "checksum-1"},
		{Version: 2, Name: "attempts", Checksum: "checksum-2"},
	}

	tests := []struct {
		name    string
		applied []AppliedMigration
		wantErr string
	}{
		{name: "empty database"},
		{
			name: "valid prefix",
			applied: []AppliedMigration{
				{Version: 1, Name: "users", Checksum: "checksum-1"},
			},
		},
		{
			name: "checksum drift",
			applied: []AppliedMigration{
				{Version: 1, Name: "users", Checksum: "changed"},
			},
			wantErr: "checksum",
		},
		{
			name: "renamed migration",
			applied: []AppliedMigration{
				{Version: 1, Name: "renamed", Checksum: "checksum-1"},
			},
			wantErr: "name",
		},
		{
			name: "out of order",
			applied: []AppliedMigration{
				{Version: 2, Name: "attempts", Checksum: "checksum-2"},
			},
			wantErr: "prefix",
		},
		{
			name: "database contains unknown version",
			applied: []AppliedMigration{
				{Version: 1, Name: "users", Checksum: "checksum-1"},
				{Version: 2, Name: "attempts", Checksum: "checksum-2"},
				{Version: 3, Name: "unknown", Checksum: "checksum-3"},
			},
			wantErr: "missing from source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateApplied(migrations, tt.applied)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateApplied() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("validateApplied() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestMigratorUpIsIdempotentAndDetectsDrift(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	schema := "migration_test_" + strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "")
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA `+quoteIdentifier(schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	defer db.ExecContext(context.Background(), `DROP SCHEMA `+quoteIdentifier(schema)+` CASCADE`)

	source := fstest.MapFS{
		"000001_users.up.sql": {Data: []byte("CREATE TABLE users (id bigint PRIMARY KEY);\n")},
	}
	migrator, err := NewMigrator(db, source, MigratorOptions{Schema: schema})
	if err != nil {
		t.Fatalf("NewMigrator() error = %v", err)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("first Up() error = %v", err)
	}
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("second Up() error = %v", err)
	}

	status, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if len(status) != 1 || status[0].State != MigrationApplied {
		t.Fatalf("Status() = %#v, want one applied migration", status)
	}

	drifted, err := NewMigrator(db, fstest.MapFS{
		"000001_users.up.sql": {Data: []byte("CREATE TABLE users (id uuid PRIMARY KEY);\n")},
	}, MigratorOptions{Schema: schema})
	if err != nil {
		t.Fatalf("NewMigrator(drifted) error = %v", err)
	}
	if err := drifted.Up(ctx); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("drifted Up() error = %v, want checksum error", err)
	}
}

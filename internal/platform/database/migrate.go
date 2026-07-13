package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const migrationAdvisoryLockID int64 = 6_767_111_001

var (
	migrationFilenamePattern = regexp.MustCompile(`^(\d{6})_([a-z0-9]+(?:_[a-z0-9]+)*)\.up\.sql$`)
	identifierPattern        = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
)

type Migration struct {
	Version  int64
	Name     string
	Checksum string
	SQL      []byte
}

type AppliedMigration struct {
	Version   int64
	Name      string
	Checksum  string
	AppliedAt time.Time
}

type MigrationState string

const (
	MigrationApplied MigrationState = "applied"
	MigrationPending MigrationState = "pending"
)

type MigrationStatus struct {
	Version   int64
	Name      string
	Checksum  string
	State     MigrationState
	AppliedAt *time.Time
}

type MigratorOptions struct {
	Schema string
}

type Migrator struct {
	db         *sql.DB
	migrations []Migration
	schema     string
}

func DiscoverMigrations(source fs.FS) ([]Migration, error) {
	if source == nil {
		return nil, errors.New("migration source is required")
	}

	var migrations []Migration
	versions := make(map[int64]string)
	err := fs.WalkDir(source, ".", func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(filePath, ".up.sql") {
			return nil
		}

		filename := path.Base(filePath)
		matches := migrationFilenamePattern.FindStringSubmatch(filename)
		if matches == nil {
			return fmt.Errorf("invalid migration filename %q: want NNNNNN_name.up.sql", filePath)
		}

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil || version <= 0 {
			return fmt.Errorf("invalid migration version in %q", filePath)
		}
		if previous, exists := versions[version]; exists {
			return fmt.Errorf("duplicate migration version %06d in %q and %q", version, previous, filePath)
		}

		contents, err := fs.ReadFile(source, filePath)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", filePath, err)
		}
		if strings.TrimSpace(string(contents)) == "" {
			return fmt.Errorf("migration %q is empty", filePath)
		}

		digest := sha256.Sum256(contents)
		migrations = append(migrations, Migration{
			Version:  version,
			Name:     matches[2],
			Checksum: hex.EncodeToString(digest[:]),
			SQL:      contents,
		})
		versions[version] = filePath
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover migrations: %w", err)
	}
	if len(migrations) == 0 {
		return nil, errors.New("no up migrations found")
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

func NewMigrator(db *sql.DB, source fs.FS, options MigratorOptions) (*Migrator, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}

	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !identifierPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}

	migrations, err := DiscoverMigrations(source)
	if err != nil {
		return nil, err
	}

	return &Migrator{db: db, migrations: migrations, schema: schema}, nil
}

func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Close()

	if err := m.prepareConnection(ctx, conn); err != nil {
		return nil, err
	}
	applied, err := loadApplied(ctx, conn)
	if err != nil {
		return nil, err
	}
	if err := validateApplied(m.migrations, applied); err != nil {
		return nil, err
	}

	return buildStatus(m.migrations, applied), nil
}

func (m *Migrator) Up(ctx context.Context) error {
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrationAdvisoryLockID); err != nil {
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.ExecContext(unlockCtx, `SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockID)
	}()

	if err := m.prepareConnection(ctx, conn); err != nil {
		return err
	}
	applied, err := loadApplied(ctx, conn)
	if err != nil {
		return err
	}
	if err := validateApplied(m.migrations, applied); err != nil {
		return err
	}

	for _, migration := range m.migrations[len(applied):] {
		if err := applyMigration(ctx, conn, migration); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) prepareConnection(ctx context.Context, conn *sql.Conn) error {
	if _, err := conn.ExecContext(ctx, `SET search_path TO `+quoteIdentifier(m.schema)); err != nil {
		return fmt.Errorf("set migration search path: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint PRIMARY KEY CHECK (version > 0),
			name text NOT NULL,
			checksum char(64) NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("bootstrap schema_migrations: %w", err)
	}
	return nil
}

func loadApplied(ctx context.Context, conn *sql.Conn) ([]AppliedMigration, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT version, name, checksum, applied_at
		FROM schema_migrations
		ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	var applied []AppliedMigration
	for rows.Next() {
		var migration AppliedMigration
		if err := rows.Scan(&migration.Version, &migration.Name, &migration.Checksum, &migration.AppliedAt); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied = append(applied, migration)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return applied, nil
}

func validateApplied(migrations []Migration, applied []AppliedMigration) error {
	if len(applied) > len(migrations) {
		return fmt.Errorf("applied migration version %06d is missing from source", applied[len(migrations)].Version)
	}

	for i, record := range applied {
		migration := migrations[i]
		if record.Version != migration.Version {
			return fmt.Errorf("applied migrations are not an exact source prefix: got version %06d at position %d, want %06d", record.Version, i+1, migration.Version)
		}
		if record.Name != migration.Name {
			return fmt.Errorf("migration %06d name changed from %q to %q", migration.Version, record.Name, migration.Name)
		}
		if record.Checksum != migration.Checksum {
			return fmt.Errorf("migration %06d checksum changed: database=%s source=%s", migration.Version, record.Checksum, migration.Checksum)
		}
	}
	return nil
}

func buildStatus(migrations []Migration, applied []AppliedMigration) []MigrationStatus {
	statuses := make([]MigrationStatus, 0, len(migrations))
	for i, migration := range migrations {
		status := MigrationStatus{
			Version:  migration.Version,
			Name:     migration.Name,
			Checksum: migration.Checksum,
			State:    MigrationPending,
		}
		if i < len(applied) {
			appliedAt := applied[i].AppliedAt
			status.State = MigrationApplied
			status.AppliedAt = &appliedAt
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func applyMigration(ctx context.Context, conn *sql.Conn, migration Migration) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %06d: %w", migration.Version, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(migration.SQL)); err != nil {
		return fmt.Errorf("apply migration %06d_%s: %w", migration.Version, migration.Name, err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, name, checksum)
		VALUES ($1, $2, $3)`, migration.Version, migration.Name, migration.Checksum); err != nil {
		return fmt.Errorf("record migration %06d: %w", migration.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %06d: %w", migration.Version, err)
	}
	return nil
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

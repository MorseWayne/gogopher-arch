package definition

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

const releaseRegistrationAdvisoryLockID int64 = 6_767_111_002

var postgresSchemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

type ReleaseStoreOptions struct {
	Schema string
}

type ReleaseStore struct {
	db     *sql.DB
	schema string
}

func NewReleaseStore(db *sql.DB, options ReleaseStoreOptions) (*ReleaseStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !postgresSchemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	return &ReleaseStore{db: db, schema: schema}, nil
}

func (s *ReleaseStore) ReferencedReleaseIDs(ctx context.Context) ([]string, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin referenced release query: %w", err)
	}
	defer tx.Rollback()
	if err := s.setSearchPath(ctx, tx); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT DISTINCT release_id
		FROM learning_attempts
		ORDER BY release_id`)
	if err != nil {
		return nil, fmt.Errorf("query referenced releases: %w", err)
	}
	defer rows.Close()
	var releaseIDs []string
	for rows.Next() {
		var releaseID string
		if err := rows.Scan(&releaseID); err != nil {
			return nil, fmt.Errorf("scan referenced release: %w", err)
		}
		releaseIDs = append(releaseIDs, releaseID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate referenced releases: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit referenced release query: %w", err)
	}
	return releaseIDs, nil
}

func (s *ReleaseStore) Register(ctx context.Context, registry *Registry) error {
	if registry == nil {
		return fmt.Errorf("definition registry is required")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin definition registration: %w", err)
	}
	defer tx.Rollback()
	if err := s.setSearchPath(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, releaseRegistrationAdvisoryLockID); err != nil {
		return fmt.Errorf("lock definition registration: %w", err)
	}

	for _, releaseID := range registry.releaseIDsForRegistration() {
		manifest, err := registry.Manifest(releaseID)
		if err != nil {
			return err
		}
		if err := registerRelease(ctx, tx, manifest); err != nil {
			return err
		}
		definitions, err := registry.Definitions(releaseID)
		if err != nil {
			return err
		}
		for _, definition := range definitions {
			if err := registerDefinition(ctx, tx, definition); err != nil {
				return err
			}
		}
	}

	current := registry.CurrentReleaseID()
	if _, err := tx.ExecContext(ctx, `
		UPDATE definition_releases
		SET status = 'superseded'
		WHERE status = 'current' AND release_id <> $1`, current); err != nil {
		return fmt.Errorf("supersede previous definition release: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE definition_releases
		SET status = 'current'
		WHERE release_id = $1`, current)
	if err != nil {
		return fmt.Errorf("mark current definition release: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("read current release update count: %w", err)
	} else if affected != 1 {
		return fmt.Errorf("current definition release %q was not registered", current)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit definition registration: %w", err)
	}
	return nil
}

func (s *ReleaseStore) CurrentReleaseID(ctx context.Context) (string, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", fmt.Errorf("begin current release query: %w", err)
	}
	defer tx.Rollback()
	if err := s.setSearchPath(ctx, tx); err != nil {
		return "", err
	}
	var releaseID string
	if err := tx.QueryRowContext(ctx, `
		SELECT release_id
		FROM definition_releases
		WHERE status = 'current'`).Scan(&releaseID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("current definition release is not registered")
		}
		return "", fmt.Errorf("query current definition release: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit current release query: %w", err)
	}
	return releaseID, nil
}

func (s *ReleaseStore) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO `+quotePostgresIdentifier(s.schema)); err != nil {
		return fmt.Errorf("set definition store search path: %w", err)
	}
	return nil
}

func registerRelease(ctx context.Context, tx *sql.Tx, manifest ReleaseManifest) error {
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("encode release %s manifest: %w", manifest.ReleaseID, err)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO definition_releases (
			release_id, schema_version, manifest, bundle_hash, status, created_at, published_at
		) VALUES ($1, $2, $3::jsonb, $4, 'superseded', $5::timestamptz, $5::timestamptz)
		ON CONFLICT (release_id) DO NOTHING`,
		manifest.ReleaseID, manifest.SchemaVersion, string(manifestJSON), manifest.BundleHash, manifest.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert definition release %s: %w", manifest.ReleaseID, err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("read release insert count: %w", err)
	} else if affected == 1 {
		return nil
	}

	var schemaVersion int
	var bundleHash string
	var manifestMatches bool
	if err := tx.QueryRowContext(ctx, `
		SELECT schema_version, bundle_hash, manifest = $2::jsonb
		FROM definition_releases
		WHERE release_id = $1
		FOR UPDATE`, manifest.ReleaseID, string(manifestJSON)).Scan(&schemaVersion, &bundleHash, &manifestMatches); err != nil {
		return fmt.Errorf("load existing definition release %s: %w", manifest.ReleaseID, err)
	}
	if schemaVersion != manifest.SchemaVersion || bundleHash != manifest.BundleHash || !manifestMatches {
		return fmt.Errorf("definition release %s conflicts with immutable database history", manifest.ReleaseID)
	}
	return nil
}

func registerDefinition(ctx context.Context, tx *sql.Tx, definition Definition) error {
	bundleHash, err := storedDefinitionBundleHash(definition)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO definition_versions (
			kind, definition_id, version, content_hash, bundle_hash, release_id, definition
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		ON CONFLICT (kind, definition_id, version) DO NOTHING`,
		definition.Kind, definition.ID, definition.Version, definition.ContentHash, bundleHash,
		definition.ReleaseID, string(definition.Document))
	if err != nil {
		return fmt.Errorf("insert %s %s version %d: %w", definition.Kind, definition.ID, definition.Version, err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("read definition insert count: %w", err)
	} else if affected == 1 {
		return nil
	}

	var contentHash string
	var storedBundleHash string
	var documentMatches bool
	if err := tx.QueryRowContext(ctx, `
		SELECT content_hash, bundle_hash, definition = $4::jsonb
		FROM definition_versions
		WHERE kind = $1 AND definition_id = $2 AND version = $3
		FOR UPDATE`, definition.Kind, definition.ID, definition.Version, string(definition.Document)).Scan(&contentHash, &storedBundleHash, &documentMatches); err != nil {
		return fmt.Errorf("load existing %s %s version %d: %w", definition.Kind, definition.ID, definition.Version, err)
	}
	if contentHash != definition.ContentHash || storedBundleHash != bundleHash || !documentMatches {
		return fmt.Errorf("%s %s version %d conflicts with immutable database history", definition.Kind, definition.ID, definition.Version)
	}
	return nil
}

func storedDefinitionBundleHash(definition Definition) (string, error) {
	switch definition.Kind {
	case KindCapability:
		return definition.ContentHash, nil
	case KindActivity:
		if definition.RuleSetHash == "" {
			return "", fmt.Errorf("activity %s version %d has no rule set hash", definition.ID, definition.Version)
		}
		return definition.RuleSetHash, nil
	case KindTask:
		if definition.BundleHash == "" {
			return "", fmt.Errorf("task %s version %d has no bundle hash", definition.ID, definition.Version)
		}
		return definition.BundleHash, nil
	default:
		return "", fmt.Errorf("unsupported definition kind %q", definition.Kind)
	}
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + identifier + `"`
}

package store

import (
	"context"
	"database/sql"
	"fmt"
)

const SchemaVersion = 1

// Execer is the subset of database/sql used by schema bootstrap.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Bootstrap creates the initial SQLite schema.
func Bootstrap(ctx context.Context, db Execer) error {
	if db == nil {
		return fmt.Errorf("database is required")
	}
	for _, stmt := range schemaStatements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER NOT NULL,
		applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS workspaces (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS corpora (
		id TEXT PRIMARY KEY,
		workspace_id TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
	)`,
	`CREATE TABLE IF NOT EXISTS sources (
		id TEXT PRIMARY KEY,
		corpus_id TEXT NOT NULL,
		name TEXT NOT NULL,
		document_hash TEXT NOT NULL,
		document_size INTEGER NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (corpus_id) REFERENCES corpora(id)
	)`,
	`CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		kind TEXT NOT NULL,
		status TEXT NOT NULL,
		target_id TEXT NOT NULL,
		error TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`INSERT INTO schema_version (version)
		SELECT 1
		WHERE NOT EXISTS (SELECT 1 FROM schema_version WHERE version = 1)`,
}

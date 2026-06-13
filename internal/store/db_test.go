package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenBootstrapsSQLiteDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "ycontext.db")

	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, table := range []string{"schema_version", "workspaces", "corpora", "sources", "jobs"} {
		var name string
		err := db.QueryRowContext(
			context.Background(),
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %s was not created: %v", table, err)
		}
		if name != table {
			t.Fatalf("table name = %q, want %q", name, table)
		}
	}
}

func TestOpenRejectsEmptyPath(t *testing.T) {
	if _, err := Open(context.Background(), ""); err == nil {
		t.Fatal("expected empty path error")
	}
}

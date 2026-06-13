package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

type recordingDB struct {
	statements []string
}

func (db *recordingDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	db.statements = append(db.statements, query)
	return nil, nil
}

func TestBootstrapExecutesInitialSchema(t *testing.T) {
	db := &recordingDB{}
	if err := Bootstrap(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	if len(db.statements) != len(schemaStatements) {
		t.Fatalf("executed %d statements, want %d", len(db.statements), len(schemaStatements))
	}
	for _, want := range []string{"schema_version", "workspaces", "corpora", "sources", "jobs"} {
		if !containsStatement(db.statements, want) {
			t.Fatalf("schema does not include %s", want)
		}
	}
}

func TestBootstrapRejectsNilDatabase(t *testing.T) {
	if err := Bootstrap(context.Background(), nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

func containsStatement(statements []string, value string) bool {
	for _, stmt := range statements {
		if strings.Contains(stmt, value) {
			return true
		}
	}
	return false
}

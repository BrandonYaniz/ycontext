package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestRepositoryStoresCorporaSourcesAndJobs(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "ycontext.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo, err := NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.CreateWorkspace(ctx, Workspace{ID: "wrk_1", Name: "default"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateCorpus(ctx, Corpus{ID: "cor_1", WorkspaceID: "wrk_1", Name: "Moby Dick"}); err != nil {
		t.Fatal(err)
	}

	corpora, err := repo.ListCorpora(ctx, "wrk_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(corpora) != 1 {
		t.Fatalf("corpora length = %d, want 1", len(corpora))
	}
	if corpora[0].ID != "cor_1" || corpora[0].Name != "Moby Dick" {
		t.Fatalf("unexpected corpus: %+v", corpora[0])
	}

	if err := repo.CreateSource(ctx, Source{
		ID:           "src_1",
		CorpusID:     "cor_1",
		Name:         "chapter-1.txt",
		DocumentHash: "7376efceaacd85bc1d8dbfdaf8a17fb7c5ce4a31d2be652a52a8e834e09c4c7e",
		DocumentSize: 17,
	}); err != nil {
		t.Fatal(err)
	}

	if err := repo.CreateJob(ctx, Job{
		ID:       "job_1",
		Kind:     "ingest",
		Status:   "queued",
		TargetID: "cor_1",
		Error:    sql.NullString{},
	}); err != nil {
		t.Fatal(err)
	}
	job, err := repo.GetJob(ctx, "job_1")
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != "queued" || job.Kind != "ingest" {
		t.Fatalf("unexpected job: %+v", job)
	}
}

func TestNewRepositoryRejectsNilDatabase(t *testing.T) {
	if _, err := NewRepository(nil); err == nil {
		t.Fatal("expected nil database error")
	}
}

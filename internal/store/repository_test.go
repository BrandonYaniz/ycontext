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
	sources, err := repo.ListSources(ctx, "cor_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 {
		t.Fatalf("sources length = %d, want 1", len(sources))
	}
	if sources[0].ID != "src_1" || sources[0].Name != "chapter-1.txt" {
		t.Fatalf("unexpected source: %+v", sources[0])
	}
	if sources[0].DocumentSize != 17 {
		t.Fatalf("document size = %d, want 17", sources[0].DocumentSize)
	}
	if err := repo.CreateNode(ctx, Node{
		ID:        "nod_1",
		SourceID:  "src_1",
		Kind:      "rough_chunk",
		Level:     0,
		Position:  0,
		StartByte: 0,
		EndByte:   8,
		Text:      "Call me",
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateNode(ctx, Node{
		ID:        "nod_2",
		SourceID:  "src_1",
		Kind:      "rough_chunk",
		Level:     0,
		Position:  1,
		StartByte: 8,
		EndByte:   17,
		Text:      "Ishmael.",
	}); err != nil {
		t.Fatal(err)
	}
	nodes, err := repo.ListSourceNodes(ctx, "src_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("nodes length = %d, want 2", len(nodes))
	}
	if nodes[0].ID != "nod_1" || nodes[0].Position != 0 || nodes[0].Text != "Call me" {
		t.Fatalf("unexpected first node: %+v", nodes[0])
	}
	if nodes[1].ID != "nod_2" || nodes[1].StartByte != 8 || nodes[1].EndByte != 17 {
		t.Fatalf("unexpected second node: %+v", nodes[1])
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

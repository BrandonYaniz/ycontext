package ingest

import (
	"context"
	"testing"

	"github.com/yanizio/ycontext/internal/store"
)

func TestIngestSourceCreatesRoughChunkNodes(t *testing.T) {
	repo := &fakeRepository{
		source: store.Source{
			ID:           "src_1",
			DocumentHash: "hash_1",
		},
	}
	docs := &fakeDocuments{
		content: []byte("one two three four five"),
	}
	service := Service{
		repository: repo,
		documents:  docs,
		newID: func(prefix string) (string, error) {
			return prefix + "_test", nil
		},
	}

	result, err := service.IngestSource(context.Background(), "src_1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceID != "src_1" || result.Chunks != 3 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if docs.hash != "hash_1" {
		t.Fatalf("document hash = %q, want hash_1", docs.hash)
	}
	if len(repo.nodes) != 3 {
		t.Fatalf("nodes length = %d, want 3", len(repo.nodes))
	}
	if repo.nodes[0].Kind != "rough_chunk" || repo.nodes[0].Position != 0 || repo.nodes[0].Text != "one two" {
		t.Fatalf("unexpected first node: %+v", repo.nodes[0])
	}
	if repo.nodes[2].Position != 2 || repo.nodes[2].Text != "five" {
		t.Fatalf("unexpected final node: %+v", repo.nodes[2])
	}

	result, err = service.IngestSource(context.Background(), "src_1", 5)
	if err != nil {
		t.Fatal(err)
	}
	if result.Chunks != 1 {
		t.Fatalf("second ingest chunks = %d, want 1", result.Chunks)
	}
	if len(repo.nodes) != 1 || repo.nodes[0].Text != "one two three four five" {
		t.Fatalf("nodes were not replaced: %+v", repo.nodes)
	}
}

func TestIngestSourceValidatesInputs(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeDocuments{})
	if _, err := service.IngestSource(context.Background(), "", 2); err == nil {
		t.Fatal("expected source id error")
	}
	if _, err := service.IngestSource(context.Background(), "src_1", 0); err == nil {
		t.Fatal("expected max words error")
	}
}

type fakeRepository struct {
	source store.Source
	nodes  []store.Node
}

func (r *fakeRepository) GetSource(ctx context.Context, id string) (store.Source, error) {
	return r.source, nil
}

func (r *fakeRepository) ReplaceRoughChunkNodes(ctx context.Context, sourceID string, nodes []store.Node) error {
	r.nodes = append([]store.Node(nil), nodes...)
	return nil
}

type fakeDocuments struct {
	hash    string
	content []byte
}

func (d *fakeDocuments) Read(ctx context.Context, hash string) ([]byte, error) {
	d.hash = hash
	return d.content, nil
}

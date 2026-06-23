package ingest

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yanizio/ycontext/internal/chunk"
	"github.com/yanizio/ycontext/internal/document"
	"github.com/yanizio/ycontext/internal/id"
	"github.com/yanizio/ycontext/internal/store"
)

type Repository interface {
	GetSource(ctx context.Context, id string) (store.Source, error)
	ReplaceRoughChunkNodes(ctx context.Context, sourceID string, nodes []store.Node) error
}

type DocumentStore interface {
	Read(ctx context.Context, hash string) ([]byte, error)
}

type IDFunc func(prefix string) (string, error)

type Service struct {
	repository Repository
	documents  DocumentStore
	newID      IDFunc
}

type Result struct {
	SourceID string
	Chunks   int
}

func NewService(repository Repository, documents DocumentStore) Service {
	return Service{
		repository: repository,
		documents:  documents,
		newID:      id.New,
	}
}

func (s Service) IngestSource(ctx context.Context, sourceID string, maxWords int) (Result, error) {
	if s.repository == nil {
		return Result{}, fmt.Errorf("repository is required")
	}
	if s.documents == nil {
		return Result{}, fmt.Errorf("document store is required")
	}
	if sourceID == "" {
		return Result{}, fmt.Errorf("source id is required")
	}

	source, err := s.repository.GetSource(ctx, sourceID)
	if err != nil {
		return Result{}, err
	}
	content, err := s.documents.Read(ctx, source.DocumentHash)
	if err != nil {
		return Result{}, err
	}
	chunks, err := chunk.Split(string(content), chunk.Options{MaxWords: maxWords})
	if err != nil {
		return Result{}, err
	}

	nodes := make([]store.Node, 0, len(chunks))
	for _, rough := range chunks {
		nodeID, err := s.makeID("nod")
		if err != nil {
			return Result{}, err
		}
		nodes = append(nodes, store.Node{
			ID:        nodeID,
			SourceID:  source.ID,
			ParentID:  sql.NullString{},
			Kind:      "rough_chunk",
			Level:     0,
			Position:  rough.Index,
			StartByte: rough.StartByte,
			EndByte:   rough.EndByte,
			Text:      rough.Text,
		})
	}
	if err := s.repository.ReplaceRoughChunkNodes(ctx, source.ID, nodes); err != nil {
		return Result{}, err
	}

	return Result{
		SourceID: source.ID,
		Chunks:   len(chunks),
	}, nil
}

func (s Service) makeID(prefix string) (string, error) {
	if s.newID == nil {
		return "", fmt.Errorf("id generator is required")
	}
	return s.newID(prefix)
}

var _ DocumentStore = document.Store{}

package daemon

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yanizio/ycontext/internal/document"
	"github.com/yanizio/ycontext/internal/id"
	"github.com/yanizio/ycontext/internal/store"
	"github.com/yanizio/ycontext/internal/version"
	"github.com/yanizio/ycontext/pkg/types"
)

type Repository interface {
	CreateWorkspace(ctx context.Context, workspace store.Workspace) error
	CreateCorpus(ctx context.Context, corpus store.Corpus) error
	CreateSource(ctx context.Context, source store.Source) error
	ListSources(ctx context.Context, corpusID string) ([]store.Source, error)
}

type DocumentStore interface {
	Put(ctx context.Context, content []byte) (document.Document, error)
}

type IDFunc func(prefix string) (string, error)

// Handler dispatches protocol requests to daemon methods.
type Handler struct {
	repository Repository
	documents  DocumentStore
	newID      IDFunc
}

func NewHandler() Handler {
	return Handler{newID: id.New}
}

func NewStorageHandler(repository Repository) Handler {
	return Handler{
		repository: repository,
		newID:      id.New,
	}
}

func NewIngestHandler(repository Repository, documents DocumentStore) Handler {
	return Handler{
		repository: repository,
		documents:  documents,
		newID:      id.New,
	}
}

func (h Handler) Handle(ctx context.Context, req types.Request) types.Response {
	if err := ctx.Err(); err != nil {
		return errorResponse(req.ID, "cancelled", err.Error())
	}
	if req.ID == "" {
		return errorResponse("", "invalid_request", "request id is required")
	}
	if req.Method == "" {
		return errorResponse(req.ID, "invalid_request", "request method is required")
	}

	switch req.Method {
	case "status":
		return types.Response{
			ID: req.ID,
			OK: true,
			Result: types.StatusResult{
				Version: version.String(),
				Ready:   true,
			},
		}
	case "workspace.create":
		return h.createWorkspace(ctx, req)
	case "corpus.create":
		return h.createCorpus(ctx, req)
	case "source.add_text":
		return h.addTextSource(ctx, req)
	case "source.list":
		return h.listSources(ctx, req)
	default:
		return errorResponse(req.ID, "method_not_found", "unknown method: "+req.Method)
	}
}

type workspaceCreateParams struct {
	Name string `json:"name"`
}

type corpusCreateParams struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
}

type sourceAddTextParams struct {
	CorpusID string `json:"corpus_id"`
	Name     string `json:"name"`
	Text     string `json:"text"`
}

type sourceListParams struct {
	CorpusID string `json:"corpus_id"`
}

func (h Handler) createWorkspace(ctx context.Context, req types.Request) types.Response {
	if h.repository == nil {
		return errorResponse(req.ID, "unavailable", "storage is not configured")
	}
	var params workspaceCreateParams
	if err := decodeParams(req.Params, &params); err != nil {
		return errorResponse(req.ID, "invalid_request", err.Error())
	}
	if params.Name == "" {
		return errorResponse(req.ID, "invalid_request", "workspace name is required")
	}

	workspaceID, err := h.makeID("wrk")
	if err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	if err := h.repository.CreateWorkspace(ctx, store.Workspace{
		ID:   workspaceID,
		Name: params.Name,
	}); err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	return types.Response{
		ID: req.ID,
		OK: true,
		Result: types.WorkspaceCreateResult{
			WorkspaceID: workspaceID,
		},
	}
}

func (h Handler) createCorpus(ctx context.Context, req types.Request) types.Response {
	if h.repository == nil {
		return errorResponse(req.ID, "unavailable", "storage is not configured")
	}
	var params corpusCreateParams
	if err := decodeParams(req.Params, &params); err != nil {
		return errorResponse(req.ID, "invalid_request", err.Error())
	}
	if params.WorkspaceID == "" {
		return errorResponse(req.ID, "invalid_request", "workspace_id is required")
	}
	if params.Name == "" {
		return errorResponse(req.ID, "invalid_request", "corpus name is required")
	}

	corpusID, err := h.makeID("cor")
	if err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	if err := h.repository.CreateCorpus(ctx, store.Corpus{
		ID:          corpusID,
		WorkspaceID: params.WorkspaceID,
		Name:        params.Name,
	}); err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	return types.Response{
		ID: req.ID,
		OK: true,
		Result: types.CorpusCreateResult{
			CorpusID: corpusID,
		},
	}
}

func (h Handler) addTextSource(ctx context.Context, req types.Request) types.Response {
	if h.repository == nil || h.documents == nil {
		return errorResponse(req.ID, "unavailable", "ingestion storage is not configured")
	}
	var params sourceAddTextParams
	if err := decodeParams(req.Params, &params); err != nil {
		return errorResponse(req.ID, "invalid_request", err.Error())
	}
	if params.CorpusID == "" {
		return errorResponse(req.ID, "invalid_request", "corpus_id is required")
	}
	if params.Name == "" {
		return errorResponse(req.ID, "invalid_request", "source name is required")
	}
	if params.Text == "" {
		return errorResponse(req.ID, "invalid_request", "source text is required")
	}

	doc, err := h.documents.Put(ctx, []byte(params.Text))
	if err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	sourceID, err := h.makeID("src")
	if err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	if err := h.repository.CreateSource(ctx, store.Source{
		ID:           sourceID,
		CorpusID:     params.CorpusID,
		Name:         params.Name,
		DocumentHash: doc.Hash,
		DocumentSize: doc.Size,
	}); err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	return types.Response{
		ID: req.ID,
		OK: true,
		Result: types.SourceAddResult{
			SourceID:     sourceID,
			DocumentHash: doc.Hash,
			DocumentSize: doc.Size,
		},
	}
}

func (h Handler) listSources(ctx context.Context, req types.Request) types.Response {
	if h.repository == nil {
		return errorResponse(req.ID, "unavailable", "storage is not configured")
	}
	var params sourceListParams
	if err := decodeParams(req.Params, &params); err != nil {
		return errorResponse(req.ID, "invalid_request", err.Error())
	}
	if params.CorpusID == "" {
		return errorResponse(req.ID, "invalid_request", "corpus_id is required")
	}
	sources, err := h.repository.ListSources(ctx, params.CorpusID)
	if err != nil {
		return errorResponse(req.ID, "internal_error", err.Error())
	}
	return types.Response{
		ID: req.ID,
		OK: true,
		Result: types.SourceListResult{
			Sources: toWireSources(sources),
		},
	}
}

func toWireSources(sources []store.Source) []types.Source {
	result := make([]types.Source, 0, len(sources))
	for _, source := range sources {
		result = append(result, types.Source{
			ID:           source.ID,
			CorpusID:     source.CorpusID,
			Name:         source.Name,
			DocumentHash: source.DocumentHash,
			DocumentSize: source.DocumentSize,
			CreatedAt:    source.CreatedAt,
		})
	}
	return result
}

func (h Handler) makeID(prefix string) (string, error) {
	if h.newID == nil {
		return "", fmt.Errorf("id generator is not configured")
	}
	return h.newID(prefix)
}

func decodeParams(params map[string]any, dst any) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return nil
}

func errorResponse(id, code, message string) types.Response {
	return types.Response{
		ID: id,
		OK: false,
		Error: &types.Error{
			Code:    code,
			Message: message,
		},
	}
}

package types

// Request is a single JSON Lines protocol request.
type Request struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// Response is the final result for a request.
type Response struct {
	ID     string `json:"id"`
	OK     bool   `json:"ok"`
	Result any    `json:"result,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

// Event is an intermediate job update for a request.
type Event struct {
	ID      string         `json:"id"`
	Event   string         `json:"event"`
	JobID   string         `json:"job_id,omitempty"`
	Stage   string         `json:"stage,omitempty"`
	Percent *int           `json:"percent,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

// Error is the wire error payload returned by the daemon.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CorpusCreateResult is the current response payload for corpus creation.
type CorpusCreateResult struct {
	CorpusID string `json:"corpus_id"`
}

// SourceAddResult is the current response payload for adding a source.
type SourceAddResult struct {
	SourceID     string `json:"source_id"`
	DocumentHash string `json:"document_hash"`
	DocumentSize int64  `json:"document_size"`
}

// Source describes source metadata returned over the wire.
type Source struct {
	ID           string `json:"id"`
	CorpusID     string `json:"corpus_id"`
	Name         string `json:"name"`
	DocumentHash string `json:"document_hash"`
	DocumentSize int64  `json:"document_size"`
	CreatedAt    string `json:"created_at"`
}

// SourceListResult is the current response payload for listing sources.
type SourceListResult struct {
	Sources []Source `json:"sources"`
}

// IngestStartResult is returned when source ingestion starts and completes.
type IngestStartResult struct {
	SourceID string `json:"source_id"`
	Chunks   int    `json:"chunks"`
}

// Node describes a stored context-tree node returned over the wire.
type Node struct {
	ID        string `json:"id"`
	SourceID  string `json:"source_id"`
	ParentID  string `json:"parent_id,omitempty"`
	Kind      string `json:"kind"`
	Level     int    `json:"level"`
	Position  int    `json:"position"`
	StartByte int    `json:"start_byte"`
	EndByte   int    `json:"end_byte"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

// NodeListResult is the current response payload for listing source nodes.
type NodeListResult struct {
	Nodes []Node `json:"nodes"`
}

// WorkspaceCreateResult is the current response payload for workspace creation.
type WorkspaceCreateResult struct {
	WorkspaceID string `json:"workspace_id"`
}

// JobStatusResult is the current response payload for job state queries.
type JobStatusResult struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// StatusResult is returned by the daemon status method.
type StatusResult struct {
	Version string `json:"version"`
	Ready   bool   `json:"ready"`
}

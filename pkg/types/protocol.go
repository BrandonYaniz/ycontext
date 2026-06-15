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

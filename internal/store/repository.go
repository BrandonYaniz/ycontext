package store

import (
	"context"
	"database/sql"
	"errors"
)

// Repository provides typed access to persisted ycontext metadata.
type Repository struct {
	db *sql.DB
}

type Workspace struct {
	ID        string
	Name      string
	CreatedAt string
}

type Corpus struct {
	ID          string
	WorkspaceID string
	Name        string
	CreatedAt   string
}

type Source struct {
	ID           string
	CorpusID     string
	Name         string
	DocumentHash string
	DocumentSize int64
	CreatedAt    string
}

type Node struct {
	ID        string
	SourceID  string
	ParentID  sql.NullString
	Kind      string
	Level     int
	Position  int
	StartByte int
	EndByte   int
	Text      string
	CreatedAt string
}

type Job struct {
	ID        string
	Kind      string
	Status    string
	TargetID  string
	Error     sql.NullString
	CreatedAt string
	UpdatedAt string
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	return &Repository{db: db}, nil
}

func (r *Repository) CreateWorkspace(ctx context.Context, workspace Workspace) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workspaces (id, name) VALUES (?, ?)`,
		workspace.ID,
		workspace.Name,
	)
	return err
}

func (r *Repository) CreateCorpus(ctx context.Context, corpus Corpus) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO corpora (id, workspace_id, name) VALUES (?, ?, ?)`,
		corpus.ID,
		corpus.WorkspaceID,
		corpus.Name,
	)
	return err
}

func (r *Repository) ListCorpora(ctx context.Context, workspaceID string) ([]Corpus, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, name, created_at
		 FROM corpora
		 WHERE workspace_id = ?
		 ORDER BY created_at, id`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var corpora []Corpus
	for rows.Next() {
		var corpus Corpus
		if err := rows.Scan(&corpus.ID, &corpus.WorkspaceID, &corpus.Name, &corpus.CreatedAt); err != nil {
			return nil, err
		}
		corpora = append(corpora, corpus)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return corpora, nil
}

func (r *Repository) CreateSource(ctx context.Context, source Source) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sources (id, corpus_id, name, document_hash, document_size)
		 VALUES (?, ?, ?, ?, ?)`,
		source.ID,
		source.CorpusID,
		source.Name,
		source.DocumentHash,
		source.DocumentSize,
	)
	return err
}

func (r *Repository) GetSource(ctx context.Context, id string) (Source, error) {
	var source Source
	err := r.db.QueryRowContext(ctx,
		`SELECT id, corpus_id, name, document_hash, document_size, created_at
		 FROM sources
		 WHERE id = ?`,
		id,
	).Scan(
		&source.ID,
		&source.CorpusID,
		&source.Name,
		&source.DocumentHash,
		&source.DocumentSize,
		&source.CreatedAt,
	)
	return source, err
}

func (r *Repository) ListSources(ctx context.Context, corpusID string) ([]Source, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, corpus_id, name, document_hash, document_size, created_at
		 FROM sources
		 WHERE corpus_id = ?
		 ORDER BY created_at, id`,
		corpusID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var source Source
		if err := rows.Scan(
			&source.ID,
			&source.CorpusID,
			&source.Name,
			&source.DocumentHash,
			&source.DocumentSize,
			&source.CreatedAt,
		); err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *Repository) CreateNode(ctx context.Context, node Node) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO nodes (id, source_id, parent_id, kind, level, position, start_byte, end_byte, text)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		node.ID,
		node.SourceID,
		node.ParentID,
		node.Kind,
		node.Level,
		node.Position,
		node.StartByte,
		node.EndByte,
		node.Text,
	)
	return err
}

func (r *Repository) ListSourceNodes(ctx context.Context, sourceID string) ([]Node, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, source_id, parent_id, kind, level, position, start_byte, end_byte, text, created_at
		 FROM nodes
		 WHERE source_id = ?
		 ORDER BY level, position, id`,
		sourceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		if err := rows.Scan(
			&node.ID,
			&node.SourceID,
			&node.ParentID,
			&node.Kind,
			&node.Level,
			&node.Position,
			&node.StartByte,
			&node.EndByte,
			&node.Text,
			&node.CreatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *Repository) CreateJob(ctx context.Context, job Job) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO jobs (id, kind, status, target_id, error)
		 VALUES (?, ?, ?, ?, ?)`,
		job.ID,
		job.Kind,
		job.Status,
		job.TargetID,
		job.Error,
	)
	return err
}

func (r *Repository) GetJob(ctx context.Context, id string) (Job, error) {
	var job Job
	err := r.db.QueryRowContext(ctx,
		`SELECT id, kind, status, target_id, error, created_at, updated_at
		 FROM jobs
		 WHERE id = ?`,
		id,
	).Scan(&job.ID, &job.Kind, &job.Status, &job.TargetID, &job.Error, &job.CreatedAt, &job.UpdatedAt)
	return job, err
}

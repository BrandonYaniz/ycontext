package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/yanizio/ycontext/pkg/types"
)

const (
	requestStatus          = "req_status"
	requestCreateWorkspace = "req_workspace_create"
	requestCreateCorpus    = "req_corpus_create"
	requestAddTextSource   = "req_source_add_text"
	requestListSources     = "req_source_list"
	requestStartIngest     = "req_ingest_start"
	requestListNodes       = "req_node_list"
)

// DialFunc matches net.Dialer.DialContext for testability.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// Client talks to ycontextd over a Unix socket.
type Client struct {
	SocketPath string
	Dial       DialFunc
}

// New returns a client bound to the given Unix socket path.
func New(socketPath string) *Client {
	return &Client{
		SocketPath: socketPath,
		Dial:       (&net.Dialer{}).DialContext,
	}
}

// Do sends one request and waits for the final response.
func (c *Client) Do(ctx context.Context, req types.Request) (types.Response, error) {
	if req.ID == "" {
		return types.Response{}, errors.New("request id is required")
	}
	if req.Method == "" {
		return types.Response{}, errors.New("request method is required")
	}
	if c == nil {
		return types.Response{}, errors.New("client is nil")
	}
	if c.Dial == nil {
		return types.Response{}, errors.New("dial function is required")
	}

	conn, err := c.Dial(ctx, "unix", c.SocketPath)
	if err != nil {
		return types.Response{}, err
	}
	defer conn.Close()

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return types.Response{}, err
	}

	var resp types.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return types.Response{}, err
	}
	if !resp.OK {
		if resp.Error == nil {
			return resp, errors.New("request failed")
		}
		return resp, fmt.Errorf("%s: %s", resp.Error.Code, resp.Error.Message)
	}
	return resp, nil
}

func (c *Client) Status(ctx context.Context) (types.StatusResult, error) {
	resp, err := c.Do(ctx, types.Request{
		ID:     requestStatus,
		Method: "status",
	})
	if err != nil {
		return types.StatusResult{}, err
	}
	var result types.StatusResult
	if err := DecodeResult(resp, &result); err != nil {
		return types.StatusResult{}, err
	}
	return result, nil
}

func (c *Client) CreateWorkspace(ctx context.Context, name string) (types.WorkspaceCreateResult, error) {
	resp, err := c.Do(ctx, types.Request{
		ID:     requestCreateWorkspace,
		Method: "workspace.create",
		Params: map[string]any{"name": name},
	})
	if err != nil {
		return types.WorkspaceCreateResult{}, err
	}
	var result types.WorkspaceCreateResult
	if err := DecodeResult(resp, &result); err != nil {
		return types.WorkspaceCreateResult{}, err
	}
	return result, nil
}

func (c *Client) CreateCorpus(ctx context.Context, workspaceID, name string) (types.CorpusCreateResult, error) {
	resp, err := c.Do(ctx, types.Request{
		ID:     requestCreateCorpus,
		Method: "corpus.create",
		Params: map[string]any{
			"workspace_id": workspaceID,
			"name":         name,
		},
	})
	if err != nil {
		return types.CorpusCreateResult{}, err
	}
	var result types.CorpusCreateResult
	if err := DecodeResult(resp, &result); err != nil {
		return types.CorpusCreateResult{}, err
	}
	return result, nil
}

func (c *Client) AddTextSource(ctx context.Context, corpusID, name, text string) (types.SourceAddResult, error) {
	resp, err := c.Do(ctx, types.Request{
		ID:     requestAddTextSource,
		Method: "source.add_text",
		Params: map[string]any{
			"corpus_id": corpusID,
			"name":      name,
			"text":      text,
		},
	})
	if err != nil {
		return types.SourceAddResult{}, err
	}
	var result types.SourceAddResult
	if err := DecodeResult(resp, &result); err != nil {
		return types.SourceAddResult{}, err
	}
	return result, nil
}

func (c *Client) ListSources(ctx context.Context, corpusID string) (types.SourceListResult, error) {
	resp, err := c.Do(ctx, types.Request{
		ID:     requestListSources,
		Method: "source.list",
		Params: map[string]any{"corpus_id": corpusID},
	})
	if err != nil {
		return types.SourceListResult{}, err
	}
	var result types.SourceListResult
	if err := DecodeResult(resp, &result); err != nil {
		return types.SourceListResult{}, err
	}
	return result, nil
}

func (c *Client) StartIngest(ctx context.Context, sourceID string, maxWords int) (types.IngestStartResult, error) {
	resp, err := c.Do(ctx, types.Request{
		ID:     requestStartIngest,
		Method: "ingest.start",
		Params: map[string]any{
			"source_id": sourceID,
			"max_words": maxWords,
		},
	})
	if err != nil {
		return types.IngestStartResult{}, err
	}
	var result types.IngestStartResult
	if err := DecodeResult(resp, &result); err != nil {
		return types.IngestStartResult{}, err
	}
	return result, nil
}

func (c *Client) ListNodes(ctx context.Context, sourceID string) (types.NodeListResult, error) {
	resp, err := c.Do(ctx, types.Request{
		ID:     requestListNodes,
		Method: "node.list",
		Params: map[string]any{"source_id": sourceID},
	})
	if err != nil {
		return types.NodeListResult{}, err
	}
	var result types.NodeListResult
	if err := DecodeResult(resp, &result); err != nil {
		return types.NodeListResult{}, err
	}
	return result, nil
}

// DecodeResult decodes the response result into dst.
func DecodeResult(resp types.Response, dst any) error {
	if dst == nil {
		return nil
	}
	raw, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}
	if string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, dst)
}

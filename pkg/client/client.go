package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/yanizio/ycontext/pkg/types"
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

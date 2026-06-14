package socket

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// ListenAndServe listens on a Unix socket path until ctx is cancelled.
func ListenAndServe(ctx context.Context, path string, handler Handler) error {
	if path == "" {
		return errors.New("socket path is required")
	}
	if handler == nil {
		return errors.New("handler is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return err
	}
	defer os.Remove(path)
	return Serve(ctx, listener, handler)
}

// Serve accepts connections from listener until ctx is cancelled or Accept fails.
func Serve(ctx context.Context, listener net.Listener, handler Handler) error {
	if listener == nil {
		return errors.New("listener is required")
	}
	if handler == nil {
		return errors.New("handler is required")
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer conn.Close()
			_ = ServeConn(ctx, conn, handler)
		}()
	}
}

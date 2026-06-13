package document

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Store keeps original source documents in a content-addressed filesystem tree.
type Store struct {
	Root string
}

// Document describes stored source content.
type Document struct {
	Hash string
	Path string
	Size int64
}

// NewStore returns a document store rooted at root.
func NewStore(root string) Store {
	return Store{Root: root}
}

// Put writes content to the store and returns its content-addressed metadata.
func (s Store) Put(ctx context.Context, content []byte) (Document, error) {
	if err := ctx.Err(); err != nil {
		return Document{}, err
	}
	if s.Root == "" {
		return Document{}, errors.New("document store root is required")
	}

	hash := hashContent(content)
	path := s.pathForHash(hash)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Document{}, err
	}

	if _, err := os.Stat(path); err == nil {
		return Document{Hash: hash, Path: path, Size: int64(len(content))}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return Document{}, err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return Document{}, err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return Document{}, err
	}
	if err := tmp.Close(); err != nil {
		return Document{}, err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return Document{}, err
	}
	cleanup = false

	return Document{Hash: hash, Path: path, Size: int64(len(content))}, nil
}

// Read returns the stored content for hash.
func (s Store) Read(ctx context.Context, hash string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.Root == "" {
		return nil, errors.New("document store root is required")
	}
	if !isSHA256(hash) {
		return nil, fmt.Errorf("invalid sha256 hash %q", hash)
	}
	return os.ReadFile(s.pathForHash(hash))
}

func (s Store) pathForHash(hash string) string {
	return filepath.Join(s.Root, "sha256", hash[:2], hash[2:4], hash)
}

func hashContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func isSHA256(hash string) bool {
	if len(hash) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}

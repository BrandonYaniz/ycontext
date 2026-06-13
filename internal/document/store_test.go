package document

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPutAndRead(t *testing.T) {
	store := NewStore(t.TempDir())
	content := []byte("Call me Ishmael.\n")

	doc, err := store.Put(context.Background(), content)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Hash == "" {
		t.Fatal("hash is empty")
	}
	if doc.Size != int64(len(content)) {
		t.Fatalf("size = %d, want %d", doc.Size, len(content))
	}
	if _, err := os.Stat(doc.Path); err != nil {
		t.Fatal(err)
	}
	if filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(doc.Path)))) != "sha256" {
		t.Fatalf("path = %q, want sha256 tree", doc.Path)
	}

	got, err := store.Read(context.Background(), doc.Hash)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("content = %q, want %q", got, content)
	}
}

func TestPutIsIdempotent(t *testing.T) {
	store := NewStore(t.TempDir())
	content := []byte("same content")

	first, err := store.Put(context.Background(), content)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Put(context.Background(), content)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("second put changed document: first=%+v second=%+v", first, second)
	}
}

func TestReadRejectsInvalidHash(t *testing.T) {
	store := NewStore(t.TempDir())
	if _, err := store.Read(context.Background(), "../bad"); err == nil {
		t.Fatal("expected invalid hash error")
	}
}

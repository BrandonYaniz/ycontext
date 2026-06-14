package id

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	first, err := New("cor")
	if err != nil {
		t.Fatal(err)
	}
	second, err := New("cor")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("ids should differ: %q", first)
	}
	if !strings.HasPrefix(first, "cor_") {
		t.Fatalf("id = %q, want cor_ prefix", first)
	}
	if len(first) != len("cor_")+32 {
		t.Fatalf("id length = %d, want %d", len(first), len("cor_")+32)
	}
}

func TestNewRejectsEmptyPrefix(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("expected empty prefix error")
	}
}

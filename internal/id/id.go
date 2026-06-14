package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// New returns a random identifier with the given short prefix.
func New(prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("id prefix is required")
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(b[:]), nil
}

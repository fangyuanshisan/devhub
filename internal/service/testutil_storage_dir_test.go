package service

import (
	"os"
	"path/filepath"
	"testing"
)

func ensureWritableTestStorageDir(t *testing.T, rel string) string {
	t.Helper()
	root := mustProjectRoot(t)
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(abs, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	return rel
}

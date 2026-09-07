package dlna

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMediaPathReady(t *testing.T) {
	if MediaPathReady("") {
		t.Fatal("empty path should be inactive")
	}
	dir := t.TempDir()
	if !MediaPathReady(dir) {
		t.Fatal("temp dir should be ready")
	}
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if MediaPathReady(file) {
		t.Fatal("file path should not be ready")
	}
}

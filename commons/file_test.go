package commons

import (
	"testing"
	"slices"

	"path/filepath"
)

func TestFileWorkspaceInfo(t *testing.T) {
	abs, err := filepath.Abs(".file/dir1/file1.json")
	if err != nil {
		t.Fatalf("%v", err)
	}

	w := FileWorkspaceInfo(abs)
	if w == nil {
		t.Fatal("should have returned")
	}

	if w.ParentDir != "dir1" {
		t.Fatalf("invalid parent dir %v", w)
	}
}

func TestFileDeepScan(t *testing.T) {
	files := make([]string, 0)

	FileDeepScan(".file", &files)

	l := len(files)

	if l != 2 {
		t.Fatalf("Should have found %d files", l)
	}

	if !slices.Contains(files, ".file/dir1/file1.json") {
		t.Fatal("Should have found file1.json")
	}
}
package commons

import (
	"log"
	"os"
	"strings"
	"path/filepath"
)

// FileDeepScan will scan dir recursively and append any files found to files.
func FileDeepScan(dir string, files *[]string) {
  entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Failed to read %v", err)
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
    if entry.IsDir() {
      FileDeepScan(fullPath, files)
    } else {
    	*files = append(*files, filepath.Join(dir, entry.Name()))
    }
  }
}

type WorkspaceInfo struct {
	ParentDir string
}

// FileWorkspaceInfo given a file will return relevant information about 
// its location, such as its parent directory.
func FileWorkspaceInfo(file string) *WorkspaceInfo {
	// 1. break down file
	clean := filepath.Clean(file)
	parts := strings.Split(clean, string(filepath.Separator))

	l := len(parts)

	if l > 3 {
		parentDir := parts[l - 2]
		return &WorkspaceInfo {
			ParentDir: parentDir,
		}
	}

	return nil
}

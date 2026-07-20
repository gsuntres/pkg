package sys

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"slices"
)

const HTML_EXT = ".html"

// FileExists check if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

// FilePathValid checks if a path contains legal characters and shapes
func FilePathValid(p string) bool {
	// 1. Check for empty inputs
	if strings.TrimSpace(p) == "" {
		return false
	}

	// 2. Clean resolves relative elements like '.', '..', and duplicate slashes
	cleaned := filepath.Clean(p)

	// 3. Block directory traversal attempts if you expect a localized file
	if strings.HasPrefix(cleaned, "..") {
		return false
	}

	return true
}

// GetHtmlFiles returns html files in a directory.
func GetHtmlFiles(dir string, ignored... string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0)

	for _, file := range files {
		if !file.IsDir() && 
			path.Ext(file.Name()) == ".html" &&
			!slices.Contains(ignored, file.Name()) {
			out = append(out, file.Name())	
		}
	}

	return out, nil
}

// GetDirs returns all directory names within the given path.
func GetDirs(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0)

	for _, file := range files {
		if file.IsDir() {
			out = append(out, file.Name())	
		}
	}

	return out, nil
}

// FileBase will return the base name of a file, that is the name without its extension.
func FileBase(file string) string {
	ext := path.Ext(file)
	base, _ := strings.CutSuffix(file, ext)

	return base
}
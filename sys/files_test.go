package sys

import "testing"

func TestFilePathValid(t *testing.T) {
	p := "/a/valid/path"

	if !FilePathValid(p) {
		t.Fatal("Should have been valid")
	}
}
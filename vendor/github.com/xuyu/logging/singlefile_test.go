package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSingleFileHandler(t *testing.T) {
	f, err := NewSingleFileHandler(filepath.Join(os.TempDir(), "sf.log"))
	if err != nil {
		t.Fatal(err)
	}
	AddHandler("file", f)
	Debug("%d, %s", 1, "OK")
}

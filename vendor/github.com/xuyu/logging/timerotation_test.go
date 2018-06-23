package logging

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestTimeRotationHandler(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	r, err := NewTimeRotationHandler(filepath.Join(os.TempDir(), "tr.log"), "060102-15:04:05")
	if err != nil {
		t.Fatal(err)
	}
	AddHandler("rotation", r)
	Debug("%d, %s", 1, "OK")
	time.Sleep(1200 * time.Millisecond)
	Info("%d, %s", 2, "OK")
	time.Sleep(1200 * time.Millisecond)
	Warning("%d, %s", 3, "OK")
	time.Sleep(1200 * time.Millisecond)
	Error("%d, %s", 4, "OK")
}

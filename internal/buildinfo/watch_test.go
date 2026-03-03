package buildinfo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testbin")

	if err := os.WriteFile(path, []byte("hello"), 0755); err != nil {
		t.Fatal(err)
	}

	h := hashFile(path)
	if h == "" {
		t.Fatal("hashFile returned empty string")
	}

	// Same content → same hash
	h2 := hashFile(path)
	if h2 != h {
		t.Errorf("same file hashed differently: %q vs %q", h, h2)
	}

	// Different content → different hash
	if err := os.WriteFile(path, []byte("world"), 0755); err != nil {
		t.Fatal(err)
	}
	h3 := hashFile(path)
	if h3 == h {
		t.Error("different content produced same hash")
	}
}

func TestHashFileMissing(t *testing.T) {
	h := hashFile("/nonexistent/path/to/binary")
	if h != "" {
		t.Errorf("hashFile of missing file = %q, want empty", h)
	}
}

func TestWatchSelfExecCallsCallback(t *testing.T) {
	dir := t.TempDir()
	fakeBin := filepath.Join(dir, "fakebin")

	// Write initial content
	if err := os.WriteFile(fakeBin, []byte("v1"), 0755); err != nil {
		t.Fatal(err)
	}

	execCalled := make(chan struct{}, 1)

	// Use the testable version with a custom exec callback
	stop := watchBinary(fakeBin, 50*time.Millisecond, func(path string, args []string, env []string) error {
		execCalled <- struct{}{}
		return nil
	})
	defer stop()

	// Overwrite the binary
	if err := os.WriteFile(fakeBin, []byte("v2"), 0755); err != nil {
		t.Fatal(err)
	}

	select {
	case <-execCalled:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("exec callback not called within timeout")
	}
}

func TestWatchSelfExecNoChangeNoCallback(t *testing.T) {
	dir := t.TempDir()
	fakeBin := filepath.Join(dir, "fakebin")

	if err := os.WriteFile(fakeBin, []byte("stable"), 0755); err != nil {
		t.Fatal(err)
	}

	execCalled := make(chan struct{}, 1)

	stop := watchBinary(fakeBin, 50*time.Millisecond, func(path string, args []string, env []string) error {
		execCalled <- struct{}{}
		return nil
	})
	defer stop()

	// Wait several intervals — should NOT trigger
	select {
	case <-execCalled:
		t.Fatal("exec callback called when binary didn't change")
	case <-time.After(300 * time.Millisecond):
		// success — no change detected
	}
}

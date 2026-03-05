package buildinfo

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"
)

// execFunc is the signature for syscall.Exec (swapped in tests).
type execFunc func(argv0 string, argv []string, envv []string) error

// WatchSelfExec hashes the running binary at startup, then re-checks every
// interval. When the on-disk binary changes (i.e. a new build was deployed),
// it calls syscall.Exec to replace the current process with the new binary.
//
// This is Unix-only. The caller should flush/close any resources that won't
// survive exec (DB connections, file locks) before this point, or use deferred
// cleanup that tolerates a missing call.
func WatchSelfExec(interval time.Duration) func() {
	self, err := os.Executable()
	if err != nil {
		log.Printf("[SelfWatch] Cannot determine executable path: %v", err)
		return func() {}
	}

	startChecksum := hashFile(self)

	return watchBinary(self, interval, os.Stdout, func(path string, args []string, env []string) error {
		fmt.Println() // newline after dots
		SetPrevChecksum(startChecksum)
		log.Printf("[SelfWatch] Binary changed on disk, re-exec-ing...")
		return syscall.Exec(path, args, os.Environ())
	})
}

// watchBinary is the testable core. It returns a stop function.
// On each tick where the binary hasn't changed, it writes a "." to heartbeat.
func watchBinary(path string, interval time.Duration, heartbeat io.Writer, exec execFunc) func() {
	startHash := hashFile(path)
	if startHash == "" {
		log.Printf("[SelfWatch] Warning: could not hash %s", path)
	}

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if current := hashFile(path); current != "" && current != startHash {
					exec(path, os.Args, os.Environ())
					return
				}
				fmt.Fprint(heartbeat, ".")
			case <-done:
				return
			}
		}
	}()

	return func() { close(done) }
}

func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)
	return fmt.Sprintf("%x", h.Sum(nil))
}

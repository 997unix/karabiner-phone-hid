package buildinfo

import (
	"fmt"
	"os"
	"runtime/debug"
)

const prevChecksumEnv = "PHONEHID_PREV_CHECKSUM"

// Identity holds the build-time and runtime identity of the binary.
type Identity struct {
	Checksum  string // SHA-256 of the running binary
	GoVersion string // Go toolchain version
	VCSCommit string // git commit hash (from Go build info)
	VCSDirty  bool   // true if built from modified working tree
}

// Self returns the Identity of the currently running binary.
func Self() Identity {
	id := Identity{}

	if exe, err := os.Executable(); err == nil {
		id.Checksum = hashFile(exe)
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		id.GoVersion = bi.GoVersion
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				id.VCSCommit = s.Value
			case "vcs.modified":
				id.VCSDirty = s.Value == "true"
			}
		}
	}

	return id
}

// ShortChecksum returns the checksum prefix used in log lines. The status
// endpoints report the same form, so a running build can be matched against a
// log line by eye.
func (id Identity) ShortChecksum() string {
	if len(id.Checksum) > 12 {
		return id.Checksum[:12]
	}
	return id.Checksum
}

// String returns a one-line summary suitable for logging.
func (id Identity) String() string {
	checksum := id.ShortChecksum()
	commit := id.VCSCommit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	dirty := ""
	if id.VCSDirty {
		dirty = "-dirty"
	}
	return fmt.Sprintf("checksum=%s git=%s%s go=%s", checksum, commit, dirty, id.GoVersion)
}

// PrevChecksum returns the checksum of the binary we re-exec'd from,
// or empty string if this is a fresh start (not a re-exec).
func PrevChecksum() string {
	return os.Getenv(prevChecksumEnv)
}

// SetPrevChecksum stores the current checksum in the environment
// so the next exec'd process can detect it was re-exec'd.
func SetPrevChecksum(checksum string) {
	os.Setenv(prevChecksumEnv, checksum)
}

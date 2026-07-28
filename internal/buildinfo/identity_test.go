package buildinfo

import (
	"os"
	"strings"
	"testing"
)

func TestSelfHasChecksum(t *testing.T) {
	id := Self()
	if id.Checksum == "" {
		t.Error("Self().Checksum is empty")
	}
}

func TestSelfHasGoVersion(t *testing.T) {
	id := Self()
	if id.GoVersion == "" {
		t.Error("Self().GoVersion is empty")
	}
	if !strings.HasPrefix(id.GoVersion, "go") {
		t.Errorf("GoVersion = %q, want go* prefix", id.GoVersion)
	}
}

func TestIdentityString(t *testing.T) {
	id := Identity{
		Checksum:  "abcdef1234567890abcdef1234567890",
		GoVersion: "go1.22.0",
		VCSCommit: "deadbeef12345",
		VCSDirty:  false,
	}

	s := id.String()
	if !strings.Contains(s, "checksum=abcdef123456") {
		t.Errorf("String() missing truncated checksum: %q", s)
	}
	if !strings.Contains(s, "git=deadbee") {
		t.Errorf("String() missing truncated commit: %q", s)
	}
	if !strings.Contains(s, "go=go1.22.0") {
		t.Errorf("String() missing go version: %q", s)
	}
	if strings.Contains(s, "dirty") {
		t.Errorf("String() has dirty when VCSDirty=false: %q", s)
	}
}

func TestIdentityStringDirty(t *testing.T) {
	id := Identity{
		Checksum:  "abcdef1234567890",
		GoVersion: "go1.22.0",
		VCSCommit: "deadbeef12345",
		VCSDirty:  true,
	}

	s := id.String()
	if !strings.Contains(s, "deadbee-dirty") {
		t.Errorf("String() missing dirty flag: %q", s)
	}
}

func TestIdentityStringShortValues(t *testing.T) {
	id := Identity{
		Checksum:  "abc",
		GoVersion: "go1.22.0",
		VCSCommit: "dead",
	}

	s := id.String()
	if !strings.Contains(s, "checksum=abc") {
		t.Errorf("Short checksum not preserved: %q", s)
	}
	if !strings.Contains(s, "git=dead") {
		t.Errorf("Short commit not preserved: %q", s)
	}
}

func TestPrevChecksumRoundTrip(t *testing.T) {
	// Clear any existing value
	os.Unsetenv(prevChecksumEnv)

	if got := PrevChecksum(); got != "" {
		t.Errorf("PrevChecksum() = %q, want empty", got)
	}

	SetPrevChecksum("abc123")
	if got := PrevChecksum(); got != "abc123" {
		t.Errorf("PrevChecksum() = %q, want abc123", got)
	}

	// Clean up
	os.Unsetenv(prevChecksumEnv)
}

func TestShortChecksumMatchesLogForm(t *testing.T) {
	id := Identity{Checksum: "91baba3a800c522bff1b3b2ddef695805b5821ac"}
	if got, want := id.ShortChecksum(), "91baba3a800c"; got != want {
		t.Errorf("ShortChecksum() = %q, want %q", got, want)
	}
	if !strings.Contains(id.String(), id.ShortChecksum()) {
		t.Error("String() should use the same short form so logs and status agree")
	}
}

func TestShortChecksumHandlesShortInput(t *testing.T) {
	id := Identity{Checksum: "abc"}
	if got := id.ShortChecksum(); got != "abc" {
		t.Errorf("ShortChecksum() = %q, want %q", got, "abc")
	}
}

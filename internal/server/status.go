package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tonyjiang/karabiner-phone-hid/internal/buildinfo"
	"github.com/tonyjiang/karabiner-phone-hid/internal/hid"
)

// SetReadinessSource wires the HID link into /readyz. Without it, /readyz
// reports on the HTTP server alone -- which is what the stub build wants,
// since it has no daemon behind it.
func (s *Server) SetReadinessSource(r hid.ReadinessReporter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readiness = r
}

func (s *Server) readinessSource() hid.ReadinessReporter {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readiness
}

// handleHealthz reports liveness: the process is up and serving.
//
// It deliberately consults nothing external. A Karabiner outage must not read
// as "this process is broken" -- the client reconnects on its own, and
// restarting the server would not help.
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	id := buildinfo.Self()
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"uptime":   s.uptime(),
		"checksum": id.ShortChecksum(),
	})
}

// handleReadyz reports whether input can actually be delivered, returning 503
// when it cannot. 503 rather than 500 or a hung socket, so a caller can tell
// "not ready yet" from "broken".
func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	id := buildinfo.Self()

	// No source configured means no daemon link to gate on.
	ready := hid.Readiness{Connected: true, KeyboardReady: true, PointingReady: true}
	if src := s.readinessSource(); src != nil {
		ready = src.Readiness()
	}

	status := http.StatusOK
	if !ready.OK() {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, map[string]any{
		"ready":          ready.OK(),
		"connected":      ready.Connected,
		"keyboard_ready": ready.KeyboardReady,
		"pointing_ready": ready.PointingReady,
		"uptime":         s.uptime(),
		"checksum":       id.ShortChecksum(),
	})
}

func (s *Server) uptime() string {
	return time.Since(s.startedAt).Round(time.Second).String()
}

func writeJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	// Status is a point-in-time answer; a cached one is worse than none.
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

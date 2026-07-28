package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/tonyjiang/karabiner-phone-hid/internal/hid"
)

type fakeReadiness struct{ r hid.Readiness }

func (f fakeReadiness) Readiness() hid.Readiness { return f.r }

func newStatusServer(t *testing.T) *Server {
	t.Helper()
	return NewServer(nil, "test", nil)
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v (body %q)", err, rec.Body.String())
	}
	return body
}

// Liveness must not depend on Karabiner. A daemon outage should never make
// a supervisor think the process needs restarting.
func TestHealthzOKWhileDaemonUnreachable(t *testing.T) {
	s := newStatusServer(t)
	s.SetReadinessSource(fakeReadiness{hid.Readiness{}})

	rec := httptest.NewRecorder()
	s.handleHealthz(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("healthz = %d, want 200 even with the daemon down", rec.Code)
	}
	if got := decodeBody(t, rec)["status"]; got != "ok" {
		t.Errorf("status = %v, want ok", got)
	}
}

func TestReadyzOKWhenFullyReady(t *testing.T) {
	s := newStatusServer(t)
	s.SetReadinessSource(fakeReadiness{hid.Readiness{
		Connected:     true,
		KeyboardReady: true,
		PointingReady: true,
	}})

	rec := httptest.NewRecorder()
	s.handleReadyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("readyz = %d, want 200", rec.Code)
	}

	body := decodeBody(t, rec)
	for _, field := range []string{"ready", "connected", "keyboard_ready", "pointing_ready"} {
		if body[field] != true {
			t.Errorf("%s = %v, want true", field, body[field])
		}
	}
	if body["checksum"] == nil {
		t.Error("checksum missing; it identifies which build answered")
	}
	if body["uptime"] == nil {
		t.Error("uptime missing")
	}
}

// 503 rather than 500 or a hung socket, so a poller can tell "not ready yet"
// from "broken".
func TestReadyzUnavailableWhenNotReady(t *testing.T) {
	tests := []struct {
		name string
		r    hid.Readiness
	}{
		{"disconnected", hid.Readiness{}},
		{"connected but no devices", hid.Readiness{Connected: true}},
		{"keyboard only", hid.Readiness{Connected: true, KeyboardReady: true}},
		{"pointing only", hid.Readiness{Connected: true, PointingReady: true}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newStatusServer(t)
			s.SetReadinessSource(fakeReadiness{tc.r})

			rec := httptest.NewRecorder()
			s.handleReadyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

			if rec.Code != http.StatusServiceUnavailable {
				t.Errorf("readyz = %d, want 503", rec.Code)
			}
			if body := decodeBody(t, rec); body["ready"] != false {
				t.Errorf("ready = %v, want false", body["ready"])
			}
		})
	}
}

// The stub build wires up no daemon link, so there is nothing to gate on and
// readiness reflects the HTTP server alone.
func TestReadyzWithoutSourceReportsReady(t *testing.T) {
	s := newStatusServer(t)

	rec := httptest.NewRecorder()
	s.handleReadyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("readyz = %d, want 200 when no HID link is configured", rec.Code)
	}
}

// Registered on the mux, not just reachable as methods.
func TestStatusRoutesRegistered(t *testing.T) {
	s := newStatusServer(t)
	s.SetReadinessSource(fakeReadiness{hid.Readiness{
		Connected: true, KeyboardReady: true, PointingReady: true,
	}})

	port, err := s.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer s.Stop()

	for _, path := range []string{"/healthz", "/readyz"} {
		resp, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port) + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, resp.StatusCode)
		}
	}
}

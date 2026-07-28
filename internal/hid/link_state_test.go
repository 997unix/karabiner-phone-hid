package hid

import "testing"

func TestConnectionTransitionReportedOnce(t *testing.T) {
	var s linkState

	if !s.setConnected(true) {
		t.Fatal("first connect should report a transition")
	}
	if s.setConnected(true) {
		t.Error("repeat connect should not report a transition")
	}
	if !s.setConnected(false) {
		t.Error("disconnect should report a transition")
	}
	if s.setConnected(false) {
		t.Error("repeat disconnect should not report a transition")
	}
}

// The daemon retries the connection once a second while it is unreachable.
// Logging every attempt is what grew the log to 8MB.
func TestRepeatedConnectFailuresReportOnce(t *testing.T) {
	var s linkState
	s.setConnected(true)

	reported := 0
	for i := 0; i < 100; i++ {
		if s.setConnected(false) {
			reported++
		}
	}
	if reported != 1 {
		t.Errorf("100 consecutive failures reported %d times, want 1", reported)
	}
}

func TestPointingReadyTransitionReportedOnce(t *testing.T) {
	var s linkState

	if !s.setPointingReady(true) {
		t.Fatal("first ready should report a transition")
	}
	if s.setPointingReady(true) {
		t.Error("repeat ready should not report a transition")
	}
	if !s.setPointingReady(false) {
		t.Error("becoming not-ready should report a transition")
	}
}

// v8 emits virtual_hid_pointing_ready(false) via clear_state() when the
// connection closes, but a dropped socket may not deliver it. Losing the
// connection must clear readiness regardless.
func TestDisconnectClearsPointingReadiness(t *testing.T) {
	var s linkState
	s.setConnected(true)
	s.setPointingReady(true)

	s.setConnected(false)

	if s.pointingIsReady() {
		t.Error("pointing should not be considered ready while disconnected")
	}
}

func TestPointingNeedsInitOnlyWhenConnectedAndNotReady(t *testing.T) {
	tests := []struct {
		name      string
		connected bool
		ready     bool
		want      bool
	}{
		{"disconnected and not ready", false, false, false},
		{"disconnected but flagged ready", false, true, false},
		{"connected and ready", true, true, false},
		{"connected but not ready", true, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var s linkState
			s.setConnected(tc.connected)
			if tc.ready {
				s.setPointingReady(true)
			}
			if got := s.pointingNeedsInit(); got != tc.want {
				t.Errorf("pointingNeedsInit() = %v, want %v", got, tc.want)
			}
		})
	}
}

// The watchdog ticks every 30s. A device that stays down must not produce a
// log line on every tick.
func TestWatchdogReportsOncePerOutage(t *testing.T) {
	var s linkState
	s.setConnected(true)

	reported := 0
	for i := 0; i < 10; i++ {
		if s.pointingNeedsInit() && s.shouldReportReinit() {
			reported++
		}
	}
	if reported != 1 {
		t.Errorf("sustained outage reported %d times, want 1", reported)
	}

	// Recovering and failing again is a new outage, worth reporting.
	s.setPointingReady(true)
	s.setPointingReady(false)

	if !(s.pointingNeedsInit() && s.shouldReportReinit()) {
		t.Error("a fresh outage after recovery should be reported")
	}
}

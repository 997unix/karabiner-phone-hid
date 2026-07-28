package hid

import "sync"

// linkState tracks the daemon connection and virtual-device readiness so that
// both logging and recovery are edge-triggered.
//
// Karabiner-DriverKit-VirtualHIDDevice 8.x reports readiness as a state change:
// virtual_hid_pointing_ready fires when the value actually changes, not on a
// timer. Earlier versions re-broadcast it periodically, which let the old
// watchdog treat it as a heartbeat and re-initialize when it went quiet. Under
// 8.x that heartbeat never arrives, so the absence of one means nothing and
// recovery has to key off the readiness state itself.
//
// The setters report whether the value changed, so callers can log transitions
// instead of repeating a line on every callback. The client retries the socket
// once a second while the daemon is unreachable, so an unconditional log there
// produces a line per second for as long as the outage lasts.
//
// The zero value is ready to use: disconnected, with no device ready.
type linkState struct {
	mu            sync.Mutex
	connected     bool
	keyboardReady bool
	pointingReady bool

	// reinitReported suppresses repeat watchdog logs within a single outage.
	reinitReported bool
}

// setConnected records the connection state and reports whether it changed.
// Losing the connection clears device readiness: the devices do not survive the
// daemon going away, and a dropped socket may never deliver the corresponding
// ready(false) callbacks.
func (s *linkState) setConnected(connected bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected == connected {
		return false
	}
	s.connected = connected
	if !connected {
		s.keyboardReady = false
		s.pointingReady = false
	}
	return true
}

// setKeyboardReady records keyboard readiness and reports whether it changed.
func (s *linkState) setKeyboardReady(ready bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.keyboardReady == ready {
		return false
	}
	s.keyboardReady = ready
	return true
}

// setPointingReady records pointing readiness and reports whether it changed.
// Becoming ready ends the current outage, so the next one is reported again.
func (s *linkState) setPointingReady(ready bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ready {
		s.reinitReported = false
	}
	if s.pointingReady == ready {
		return false
	}
	s.pointingReady = ready
	return true
}

// pointingIsReady reports whether the pointing device is usable.
func (s *linkState) pointingIsReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.pointingReady
}

// pointingNeedsInit reports whether the pointing device should be
// re-initialized. Only meaningful while connected: the client reconnects on its
// own, and initialize requests sent while the socket is down go nowhere.
func (s *linkState) pointingNeedsInit() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.connected && !s.pointingReady
}

// shouldReportReinit reports whether this recovery attempt is the first of the
// current outage, so a device that stays down logs once rather than on every
// watchdog tick.
func (s *linkState) shouldReportReinit() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reinitReported {
		return false
	}
	s.reinitReported = true
	return true
}

package hid

// Readiness describes the live state of the link to the Karabiner daemon.
//
// This is the answer to "is it working right now", which the log cannot give.
// s6-log is size-bounded, so startup lines age out, and readiness is logged
// only on transitions -- a healthy server that has been up for days writes
// nothing at all. Silence in the log is therefore ambiguous; this is not.
type Readiness struct {
	Connected     bool `json:"connected"`
	KeyboardReady bool `json:"keyboard_ready"`
	PointingReady bool `json:"pointing_ready"`
}

// OK reports whether input can actually be delivered. All three must hold: a
// connection with no devices behind it accepts keystrokes and drops them.
func (r Readiness) OK() bool {
	return r.Connected && r.KeyboardReady && r.PointingReady
}

// ReadinessReporter is implemented by posters backed by a live daemon
// connection. MockPoster deliberately does not implement it -- there is no
// daemon behind it to report on.
type ReadinessReporter interface {
	Readiness() Readiness
}

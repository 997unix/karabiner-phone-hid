package server

import (
	"encoding/json"
	"testing"

	"github.com/tonyjiang/karabiner-phone-hid/internal/hid"
	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

// mockResolver implements ActionResolver for tests.
type mockResolver struct {
	actions map[string][]protocol.KeyStep
}

func (m *mockResolver) Resolve(name string) ([]protocol.KeyStep, bool) {
	steps, ok := m.actions[name]
	return steps, ok
}

func newTestRouter() (*Router, *hid.MockPoster) {
	mock := &hid.MockPoster{}
	dispatcher := hid.NewDispatcher(mock)
	resolver := &mockResolver{
		actions: map[string][]protocol.KeyStep{
			"superwhisper_toggle": {
				{Key: "spacebar", Modifiers: []string{"option"}},
			},
			"tmux_copy_mode": {
				{Key: "backslash", Modifiers: []string{"control"}, DelayMs: 100},
				{Key: "open_bracket", Modifiers: []string{}},
			},
		},
	}
	return NewRouter(dispatcher, resolver), mock
}

func decodeAck(t *testing.T, data []byte) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal ack: %v", err)
	}
	return result
}

func TestRouteKeypressAction(t *testing.T) {
	router, mock := newTestRouter()

	raw := `{"type":"action","id":"k-1","action":"keypress","payload":{"key":"return_or_enter","modifiers":[]}}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}
	if ack["id"] != "k-1" {
		t.Errorf("id = %v, want k-1", ack["id"])
	}

	if len(mock.Calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(mock.Calls))
	}
	if mock.Calls[0].Keys[0] != 0x28 { // return_or_enter
		t.Errorf("key = 0x%02X, want 0x28", mock.Calls[0].Keys[0])
	}
}

func TestRouteKeypressWithModifiers(t *testing.T) {
	router, mock := newTestRouter()

	raw := `{"type":"action","id":"k-2","action":"keypress","payload":{"key":"a","modifiers":["shift","command"]}}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}

	wantMod := uint8(0x02 | 0x08) // shift | command
	if mock.Calls[0].Modifiers != wantMod {
		t.Errorf("modifiers = 0x%02X, want 0x%02X", mock.Calls[0].Modifiers, wantMod)
	}
}

func TestRouteSequenceAction(t *testing.T) {
	router, mock := newTestRouter()

	raw := `{"type":"action","id":"s-1","action":"sequence","payload":{"steps":[{"key":"backslash","modifiers":["control"]},{"key":"open_bracket","modifiers":[]}]}}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}

	// 2 steps × 2 calls = 4
	if len(mock.Calls) != 4 {
		t.Fatalf("calls = %d, want 4", len(mock.Calls))
	}
}

func TestRouteNamedAction(t *testing.T) {
	router, mock := newTestRouter()

	raw := `{"type":"action","id":"n-1","action":"superwhisper_toggle","payload":{}}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}

	// superwhisper_toggle = option+spacebar → 2 calls
	if len(mock.Calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(mock.Calls))
	}
	if mock.Calls[0].Modifiers != 0x04 { // option
		t.Errorf("modifiers = 0x%02X, want 0x04", mock.Calls[0].Modifiers)
	}
	if mock.Calls[0].Keys[0] != 0x2C { // spacebar
		t.Errorf("key = 0x%02X, want 0x2C", mock.Calls[0].Keys[0])
	}
}

func TestRoutePointingAction(t *testing.T) {
	router, mock := newTestRouter()

	raw := `{"type":"action","id":"p-1","action":"pointing","payload":{"buttons":0,"x":5,"y":-3,"vertical_wheel":0,"horizontal_wheel":0}}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}
	if ack["id"] != "p-1" {
		t.Errorf("id = %v, want p-1", ack["id"])
	}

	if len(mock.Calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(mock.Calls))
	}
	if !mock.Calls[0].Pointing {
		t.Error("call[0] should be pointing")
	}
	if mock.Calls[0].X != 5 {
		t.Errorf("x = %d, want 5", mock.Calls[0].X)
	}
	if mock.Calls[0].Y != -3 {
		t.Errorf("y = %d, want -3", mock.Calls[0].Y)
	}
}

func TestRoutePointingWithScroll(t *testing.T) {
	router, mock := newTestRouter()

	raw := `{"type":"action","id":"p-2","action":"pointing","payload":{"buttons":0,"x":0,"y":0,"vertical_wheel":-2,"horizontal_wheel":0}}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}

	if mock.Calls[0].VWheel != -2 {
		t.Errorf("vwheel = %d, want -2", mock.Calls[0].VWheel)
	}
}

func TestRouteInvalidPointingPayload(t *testing.T) {
	router, _ := newTestRouter()

	raw := `{"type":"action","id":"bad-p","action":"pointing","payload":"not an object"}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "error" {
		t.Errorf("status = %v, want error", ack["status"])
	}
}

func TestRouteUnknownAction(t *testing.T) {
	router, _ := newTestRouter()

	raw := `{"type":"action","id":"u-1","action":"nonexistent","payload":{}}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "error" {
		t.Errorf("status = %v, want error", ack["status"])
	}
	if ack["id"] != "u-1" {
		t.Errorf("id = %v, want u-1", ack["id"])
	}
}

func TestRouteInvalidJSON(t *testing.T) {
	router, _ := newTestRouter()

	resp := router.Route([]byte("not json"))

	ack := decodeAck(t, resp)
	if ack["status"] != "error" {
		t.Errorf("status = %v, want error", ack["status"])
	}
}

func TestRouteUnknownMessageType(t *testing.T) {
	router, _ := newTestRouter()

	raw := `{"type":"ping","id":"p-1"}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "error" {
		t.Errorf("status = %v, want error", ack["status"])
	}
}

func TestRouteInvalidKeypressPayload(t *testing.T) {
	router, _ := newTestRouter()

	raw := `{"type":"action","id":"bad-1","action":"keypress","payload":"not an object"}`
	resp := router.Route([]byte(raw))

	ack := decodeAck(t, resp)
	if ack["status"] != "error" {
		t.Errorf("status = %v, want error", ack["status"])
	}
}

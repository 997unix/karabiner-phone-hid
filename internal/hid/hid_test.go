package hid

import (
	"testing"

	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

// --- Key Code Tests ---

func TestLookupLetters(t *testing.T) {
	tests := []struct {
		name string
		want uint16
	}{
		{"a", 0x04}, {"b", 0x05}, {"z", 0x1D},
	}
	for _, tt := range tests {
		code, ok := LookupKeyCode(tt.name)
		if !ok {
			t.Errorf("LookupKeyCode(%q) not found", tt.name)
			continue
		}
		if code != tt.want {
			t.Errorf("LookupKeyCode(%q) = 0x%02X, want 0x%02X", tt.name, code, tt.want)
		}
	}
}

func TestLookupNumbers(t *testing.T) {
	code, ok := LookupKeyCode("1")
	if !ok || code != 0x1E {
		t.Errorf("LookupKeyCode(\"1\") = 0x%02X, %v; want 0x1E, true", code, ok)
	}
	code, ok = LookupKeyCode("0")
	if !ok || code != 0x27 {
		t.Errorf("LookupKeyCode(\"0\") = 0x%02X, %v; want 0x27, true", code, ok)
	}
}

func TestLookupSpecialKeys(t *testing.T) {
	tests := []struct {
		name string
		want uint16
	}{
		{"return_or_enter", 0x28},
		{"escape", 0x29},
		{"spacebar", 0x2C},
		{"tab", 0x2B},
		{"backslash", 0x31},
		{"open_bracket", 0x2F},
	}
	for _, tt := range tests {
		code, ok := LookupKeyCode(tt.name)
		if !ok {
			t.Errorf("LookupKeyCode(%q) not found", tt.name)
			continue
		}
		if code != tt.want {
			t.Errorf("LookupKeyCode(%q) = 0x%02X, want 0x%02X", tt.name, code, tt.want)
		}
	}
}

func TestLookupNumpadKeys(t *testing.T) {
	tests := []struct {
		name string
		want uint16
	}{
		{"numpad_0", 0x62}, {"numpad_1", 0x59}, {"numpad_9", 0x61},
		{"numpad_add", 0x57}, {"numpad_subtract", 0x56},
		{"numpad_multiply", 0x55}, {"numpad_divide", 0x54},
		{"numpad_enter", 0x58}, {"numpad_decimal", 0x63},
		{"insert", 0x49}, {"num_lock", 0x53},
	}
	for _, tt := range tests {
		code, ok := LookupKeyCode(tt.name)
		if !ok {
			t.Errorf("LookupKeyCode(%q) not found", tt.name)
			continue
		}
		if code != tt.want {
			t.Errorf("LookupKeyCode(%q) = 0x%02X, want 0x%02X", tt.name, code, tt.want)
		}
	}
}

func TestLookupUnknownKey(t *testing.T) {
	_, ok := LookupKeyCode("nonexistent")
	if ok {
		t.Error("LookupKeyCode(\"nonexistent\") should return false")
	}
}

func TestLookupModifier(t *testing.T) {
	tests := []struct {
		name string
		want uint8
	}{
		{"control", 0x01},
		{"shift", 0x02},
		{"option", 0x04},
		{"command", 0x08},
	}
	for _, tt := range tests {
		flag, ok := LookupModifier(tt.name)
		if !ok {
			t.Errorf("LookupModifier(%q) not found", tt.name)
			continue
		}
		if flag != tt.want {
			t.Errorf("LookupModifier(%q) = 0x%02X, want 0x%02X", tt.name, flag, tt.want)
		}
	}
}

func TestBuildModifierByte(t *testing.T) {
	result := BuildModifierByte([]string{"shift", "command"})
	want := uint8(0x02 | 0x08) // shift | command = 0x0A
	if result != want {
		t.Errorf("BuildModifierByte = 0x%02X, want 0x%02X", result, want)
	}
}

func TestBuildModifierByteEmpty(t *testing.T) {
	result := BuildModifierByte([]string{})
	if result != 0 {
		t.Errorf("BuildModifierByte(empty) = 0x%02X, want 0x00", result)
	}
}

func TestBuildModifierByteIgnoresUnknown(t *testing.T) {
	result := BuildModifierByte([]string{"shift", "unknown_mod"})
	if result != 0x02 {
		t.Errorf("BuildModifierByte = 0x%02X, want 0x02", result)
	}
}

// --- Dispatcher Tests ---

func TestDispatchSingleKeypress(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	err := d.Dispatch([]protocol.KeyStep{
		{Key: "a", Modifiers: []string{}},
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Fatalf("calls = %d, want 2 (PostKeyboard + ReleaseAll)", len(mock.Calls))
	}

	// First call: PostKeyboard with 'a'
	if mock.Calls[0].Release {
		t.Error("call[0] should not be a release")
	}
	if mock.Calls[0].Modifiers != 0 {
		t.Errorf("call[0].Modifiers = 0x%02X, want 0x00", mock.Calls[0].Modifiers)
	}
	if len(mock.Calls[0].Keys) != 1 || mock.Calls[0].Keys[0] != 0x04 {
		t.Errorf("call[0].Keys = %v, want [0x04]", mock.Calls[0].Keys)
	}

	// Second call: ReleaseAll
	if !mock.Calls[1].Release {
		t.Error("call[1] should be a release")
	}
}

func TestDispatchKeypressWithModifiers(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	err := d.Dispatch([]protocol.KeyStep{
		{Key: "spacebar", Modifiers: []string{"option"}},
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(mock.Calls))
	}
	if mock.Calls[0].Modifiers != 0x04 { // option
		t.Errorf("modifiers = 0x%02X, want 0x04", mock.Calls[0].Modifiers)
	}
	if mock.Calls[0].Keys[0] != 0x2C { // spacebar
		t.Errorf("keys[0] = 0x%02X, want 0x2C", mock.Calls[0].Keys[0])
	}
}

func TestDispatchMultipleModifiers(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	err := d.Dispatch([]protocol.KeyStep{
		{Key: "v", Modifiers: []string{"command", "shift"}},
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	wantMod := uint8(0x08 | 0x02) // command | shift
	if mock.Calls[0].Modifiers != wantMod {
		t.Errorf("modifiers = 0x%02X, want 0x%02X", mock.Calls[0].Modifiers, wantMod)
	}
}

func TestDispatchSequence(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	err := d.Dispatch([]protocol.KeyStep{
		{Key: "backslash", Modifiers: []string{"control"}},
		{Key: "open_bracket", Modifiers: []string{}},
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	// 2 steps × 2 calls each = 4
	if len(mock.Calls) != 4 {
		t.Fatalf("calls = %d, want 4", len(mock.Calls))
	}

	// Step 1: backslash with control
	if mock.Calls[0].Modifiers != 0x01 {
		t.Errorf("step1 modifiers = 0x%02X, want 0x01", mock.Calls[0].Modifiers)
	}
	if mock.Calls[0].Keys[0] != 0x31 { // backslash
		t.Errorf("step1 key = 0x%02X, want 0x31", mock.Calls[0].Keys[0])
	}

	// Step 2: open_bracket no modifiers
	if mock.Calls[2].Modifiers != 0 {
		t.Errorf("step2 modifiers = 0x%02X, want 0x00", mock.Calls[2].Modifiers)
	}
	if mock.Calls[2].Keys[0] != 0x2F { // open_bracket
		t.Errorf("step2 key = 0x%02X, want 0x2F", mock.Calls[2].Keys[0])
	}
}

func TestDispatchUnknownKeyReturnsError(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	err := d.Dispatch([]protocol.KeyStep{
		{Key: "nonexistent", Modifiers: []string{}},
	})
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
	if len(mock.Calls) != 0 {
		t.Errorf("should not have made any calls, got %d", len(mock.Calls))
	}
}

func TestDispatchEmptySteps(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	err := d.Dispatch([]protocol.KeyStep{})
	if err != nil {
		t.Fatalf("Dispatch empty: %v", err)
	}
	if len(mock.Calls) != 0 {
		t.Errorf("calls = %d, want 0", len(mock.Calls))
	}
}

func TestDispatchKeyDown(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	kp := &protocol.KeypressPayload{Key: "tab", Modifiers: []string{"command"}}
	err := d.DispatchKeyDown(kp)
	if err != nil {
		t.Fatalf("DispatchKeyDown: %v", err)
	}

	if len(mock.Calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(mock.Calls))
	}
	if mock.Calls[0].Release {
		t.Error("keydown should not release")
	}
	if mock.Calls[0].Modifiers != 0x08 { // command
		t.Errorf("modifiers = 0x%02X, want 0x08", mock.Calls[0].Modifiers)
	}
	if mock.Calls[0].Keys[0] != 0x2B { // tab
		t.Errorf("key = 0x%02X, want 0x2B", mock.Calls[0].Keys[0])
	}
}

func TestDispatchKeyDownModifiersOnly(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	kp := &protocol.KeypressPayload{Key: "", Modifiers: []string{"command"}}
	err := d.DispatchKeyDown(kp)
	if err != nil {
		t.Fatalf("DispatchKeyDown: %v", err)
	}

	if len(mock.Calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(mock.Calls))
	}
	if mock.Calls[0].Modifiers != 0x08 {
		t.Errorf("modifiers = 0x%02X, want 0x08", mock.Calls[0].Modifiers)
	}
	if len(mock.Calls[0].Keys) != 0 {
		t.Errorf("keys = %v, want empty", mock.Calls[0].Keys)
	}
}

func TestDispatchKeyUp(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	err := d.DispatchKeyUp()
	if err != nil {
		t.Fatalf("DispatchKeyUp: %v", err)
	}

	if len(mock.Calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(mock.Calls))
	}
	if !mock.Calls[0].Release {
		t.Error("keyup should be a release")
	}
}

func TestDispatchPointing(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	pt := &protocol.PointingPayload{
		Buttons:         0,
		X:               10,
		Y:               -5,
		VerticalWheel:   0,
		HorizontalWheel: 0,
	}
	err := d.DispatchPointing(pt)
	if err != nil {
		t.Fatalf("DispatchPointing: %v", err)
	}

	if len(mock.Calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(mock.Calls))
	}
	if !mock.Calls[0].Pointing {
		t.Error("call[0] should be pointing")
	}
	if mock.Calls[0].X != 10 {
		t.Errorf("x = %d, want 10", mock.Calls[0].X)
	}
	if mock.Calls[0].Y != -5 {
		t.Errorf("y = %d, want -5", mock.Calls[0].Y)
	}
}

func TestDispatchPointingWithButtons(t *testing.T) {
	mock := &MockPoster{}
	d := NewDispatcher(mock)

	pt := &protocol.PointingPayload{
		Buttons:         1,
		X:               0,
		Y:               0,
		VerticalWheel:   -2,
		HorizontalWheel: 3,
	}
	err := d.DispatchPointing(pt)
	if err != nil {
		t.Fatalf("DispatchPointing: %v", err)
	}

	if mock.Calls[0].Buttons != 1 {
		t.Errorf("buttons = %d, want 1", mock.Calls[0].Buttons)
	}
	if mock.Calls[0].VWheel != -2 {
		t.Errorf("vwheel = %d, want -2", mock.Calls[0].VWheel)
	}
	if mock.Calls[0].HWheel != 3 {
		t.Errorf("hwheel = %d, want 3", mock.Calls[0].HWheel)
	}
}

func TestMockPosterReset(t *testing.T) {
	mock := &MockPoster{}
	mock.PostKeyboard(0, []uint16{0x04})
	mock.Reset()
	if len(mock.Calls) != 0 {
		t.Errorf("after Reset, calls = %d, want 0", len(mock.Calls))
	}
}

package protocol

import (
	"encoding/json"
	"testing"
)

func TestDecodeKeypressAction(t *testing.T) {
	raw := `{"type":"action","id":"abc-123","action":"keypress","payload":{"key":"return_or_enter","modifiers":[]}}`

	var msg IncomingMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if msg.Type != "action" {
		t.Errorf("type = %q, want %q", msg.Type, "action")
	}
	if msg.ID != "abc-123" {
		t.Errorf("id = %q, want %q", msg.ID, "abc-123")
	}
	if msg.Action != "keypress" {
		t.Errorf("action = %q, want %q", msg.Action, "keypress")
	}

	kp, err := msg.ParseKeypress()
	if err != nil {
		t.Fatalf("ParseKeypress: %v", err)
	}
	if kp.Key != "return_or_enter" {
		t.Errorf("key = %q, want %q", kp.Key, "return_or_enter")
	}
	if len(kp.Modifiers) != 0 {
		t.Errorf("modifiers = %v, want empty", kp.Modifiers)
	}
}

func TestDecodeKeypressWithModifiers(t *testing.T) {
	raw := `{"type":"action","id":"x","action":"keypress","payload":{"key":"a","modifiers":["shift","command"]}}`

	var msg IncomingMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	kp, err := msg.ParseKeypress()
	if err != nil {
		t.Fatalf("ParseKeypress: %v", err)
	}
	if kp.Key != "a" {
		t.Errorf("key = %q, want %q", kp.Key, "a")
	}
	if len(kp.Modifiers) != 2 {
		t.Fatalf("modifiers len = %d, want 2", len(kp.Modifiers))
	}
	if kp.Modifiers[0] != "shift" {
		t.Errorf("modifiers[0] = %q, want %q", kp.Modifiers[0], "shift")
	}
	if kp.Modifiers[1] != "command" {
		t.Errorf("modifiers[1] = %q, want %q", kp.Modifiers[1], "command")
	}
}

func TestDecodeSequenceAction(t *testing.T) {
	raw := `{"type":"action","id":"seq-1","action":"sequence","payload":{"steps":[{"key":"backslash","modifiers":["control"],"delay_ms":100},{"key":"open_bracket","modifiers":[]}]}}`

	var msg IncomingMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if msg.Action != "sequence" {
		t.Errorf("action = %q, want %q", msg.Action, "sequence")
	}

	seq, err := msg.ParseSequence()
	if err != nil {
		t.Fatalf("ParseSequence: %v", err)
	}
	if len(seq.Steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(seq.Steps))
	}
	if seq.Steps[0].Key != "backslash" {
		t.Errorf("steps[0].key = %q, want %q", seq.Steps[0].Key, "backslash")
	}
	if seq.Steps[0].DelayMs != 100 {
		t.Errorf("steps[0].delay_ms = %d, want 100", seq.Steps[0].DelayMs)
	}
	if len(seq.Steps[0].Modifiers) != 1 || seq.Steps[0].Modifiers[0] != "control" {
		t.Errorf("steps[0].modifiers = %v, want [control]", seq.Steps[0].Modifiers)
	}
	if seq.Steps[1].Key != "open_bracket" {
		t.Errorf("steps[1].key = %q, want %q", seq.Steps[1].Key, "open_bracket")
	}
	if seq.Steps[1].DelayMs != 0 {
		t.Errorf("steps[1].delay_ms = %d, want 0", seq.Steps[1].DelayMs)
	}
}

func TestDecodeNamedAction(t *testing.T) {
	raw := `{"type":"action","id":"n-1","action":"superwhisper_toggle","payload":{}}`

	var msg IncomingMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if msg.Action != "superwhisper_toggle" {
		t.Errorf("action = %q, want %q", msg.Action, "superwhisper_toggle")
	}
}

func TestEncodeAckOK(t *testing.T) {
	ack := NewAckOK("abc-123")
	data, err := json.Marshal(ack)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["type"] != "ack" {
		t.Errorf("type = %v, want %q", decoded["type"], "ack")
	}
	if decoded["id"] != "abc-123" {
		t.Errorf("id = %v, want %q", decoded["id"], "abc-123")
	}
	if decoded["status"] != "ok" {
		t.Errorf("status = %v, want %q", decoded["status"], "ok")
	}
	if _, hasError := decoded["error"]; hasError {
		t.Error("ack ok should not have error field")
	}
}

func TestEncodeAckError(t *testing.T) {
	ack := NewAckError("err-1", "unknown action: foo")
	data, err := json.Marshal(ack)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["type"] != "ack" {
		t.Errorf("type = %v, want %q", decoded["type"], "ack")
	}
	if decoded["status"] != "error" {
		t.Errorf("status = %v, want %q", decoded["status"], "error")
	}
	if decoded["error"] != "unknown action: foo" {
		t.Errorf("error = %v, want %q", decoded["error"], "unknown action: foo")
	}
}

func TestEncodeConfigMessage(t *testing.T) {
	cfg := NewConfigMessage("Tony's Mac", []ActionInfo{
		{Name: "superwhisper_toggle", Label: "SuperWhisper"},
		{Name: "return", Label: "Return"},
	})
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["type"] != "config" {
		t.Errorf("type = %v, want %q", decoded["type"], "config")
	}

	payload, ok := decoded["payload"].(map[string]interface{})
	if !ok {
		t.Fatal("payload is not an object")
	}
	if payload["server_name"] != "Tony's Mac" {
		t.Errorf("server_name = %v, want %q", payload["server_name"], "Tony's Mac")
	}

	actions, ok := payload["actions"].([]interface{})
	if !ok {
		t.Fatal("actions is not an array")
	}
	if len(actions) != 2 {
		t.Fatalf("actions len = %d, want 2", len(actions))
	}
}

func TestRoundTripKeypress(t *testing.T) {
	original := IncomingMessage{
		Type:   "action",
		ID:     "rt-1",
		Action: "keypress",
	}
	payload := KeypressPayload{Key: "spacebar", Modifiers: []string{"option"}}
	payloadBytes, _ := json.Marshal(payload)
	original.Payload = payloadBytes

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded IncomingMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Type != original.Type {
		t.Errorf("type = %q, want %q", decoded.Type, original.Type)
	}
	if decoded.ID != original.ID {
		t.Errorf("id = %q, want %q", decoded.ID, original.ID)
	}

	kp, err := decoded.ParseKeypress()
	if err != nil {
		t.Fatalf("ParseKeypress: %v", err)
	}
	if kp.Key != "spacebar" {
		t.Errorf("key = %q, want %q", kp.Key, "spacebar")
	}
}

func TestDecodePointingAction(t *testing.T) {
	raw := `{"type":"action","id":"p-1","action":"pointing","payload":{"buttons":0,"x":5,"y":-3,"vertical_wheel":0,"horizontal_wheel":0}}`

	var msg IncomingMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if msg.Action != "pointing" {
		t.Errorf("action = %q, want %q", msg.Action, "pointing")
	}

	pt, err := msg.ParsePointing()
	if err != nil {
		t.Fatalf("ParsePointing: %v", err)
	}
	if pt.Buttons != 0 {
		t.Errorf("buttons = %d, want 0", pt.Buttons)
	}
	if pt.X != 5 {
		t.Errorf("x = %d, want 5", pt.X)
	}
	if pt.Y != -3 {
		t.Errorf("y = %d, want -3", pt.Y)
	}
}

func TestDecodePointingWithButtons(t *testing.T) {
	raw := `{"type":"action","id":"p-2","action":"pointing","payload":{"buttons":1,"x":0,"y":0,"vertical_wheel":-2,"horizontal_wheel":3}}`

	var msg IncomingMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	pt, err := msg.ParsePointing()
	if err != nil {
		t.Fatalf("ParsePointing: %v", err)
	}
	if pt.Buttons != 1 {
		t.Errorf("buttons = %d, want 1", pt.Buttons)
	}
	if pt.VerticalWheel != -2 {
		t.Errorf("vertical_wheel = %d, want -2", pt.VerticalWheel)
	}
	if pt.HorizontalWheel != 3 {
		t.Errorf("horizontal_wheel = %d, want 3", pt.HorizontalWheel)
	}
}

func TestRoundTripSequence(t *testing.T) {
	steps := []KeyStep{
		{Key: "a", Modifiers: []string{"shift"}, DelayMs: 50},
		{Key: "b", Modifiers: []string{}},
	}
	payload := SequencePayload{Steps: steps}
	payloadBytes, _ := json.Marshal(payload)

	original := IncomingMessage{
		Type:    "action",
		ID:      "rt-2",
		Action:  "sequence",
		Payload: payloadBytes,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded IncomingMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	seq, err := decoded.ParseSequence()
	if err != nil {
		t.Fatalf("ParseSequence: %v", err)
	}
	if len(seq.Steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(seq.Steps))
	}
	if seq.Steps[0].Key != "a" {
		t.Errorf("steps[0].key = %q, want %q", seq.Steps[0].Key, "a")
	}
	if seq.Steps[0].DelayMs != 50 {
		t.Errorf("steps[0].delay_ms = %d, want 50", seq.Steps[0].DelayMs)
	}
}

package protocol

import "encoding/json"

// IncomingMessage represents a message from the phone.
type IncomingMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// KeypressPayload is the payload for a "keypress" action.
type KeypressPayload struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers"`
}

// SequencePayload is the payload for a "sequence" action.
type SequencePayload struct {
	Steps []KeyStep `json:"steps"`
}

// KeyStep is a single key press within a sequence.
type KeyStep struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers"`
	DelayMs   int      `json:"delay_ms,omitempty"`
}

// OutgoingMessage represents a message sent to the phone.
// Use NewAckOK, NewAckError, or NewConfigMessage to create instances.
type OutgoingMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id,omitempty"`
	Status  string      `json:"status,omitempty"`
	Error   string      `json:"error,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// ConfigPayload is sent to the phone on connect.
type ConfigPayload struct {
	ServerName string       `json:"server_name"`
	Actions    []ActionInfo `json:"actions"`
}

// ActionInfo describes an available action.
type ActionInfo struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// NewAckOK creates a successful ack response.
func NewAckOK(id string) OutgoingMessage {
	return OutgoingMessage{Type: "ack", ID: id, Status: "ok"}
}

// NewAckError creates an error ack response.
func NewAckError(id string, errMsg string) OutgoingMessage {
	return OutgoingMessage{Type: "ack", ID: id, Status: "error", Error: errMsg}
}

// NewConfigMessage creates a config message.
func NewConfigMessage(serverName string, actions []ActionInfo) OutgoingMessage {
	return OutgoingMessage{
		Type: "config",
		Payload: ConfigPayload{
			ServerName: serverName,
			Actions:    actions,
		},
	}
}

// PointingPayload is the payload for a "pointing" action.
type PointingPayload struct {
	Buttons         uint32 `json:"buttons"`
	X               int8   `json:"x"`
	Y               int8   `json:"y"`
	VerticalWheel   int8   `json:"vertical_wheel"`
	HorizontalWheel int8   `json:"horizontal_wheel"`
}

// ParsePayload decodes the raw payload into the appropriate type based on action.
func (m *IncomingMessage) ParseKeypress() (*KeypressPayload, error) {
	var p KeypressPayload
	if err := json.Unmarshal(m.Payload, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ParseSequence decodes the raw payload as a sequence.
func (m *IncomingMessage) ParseSequence() (*SequencePayload, error) {
	var p SequencePayload
	if err := json.Unmarshal(m.Payload, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ParsePointing decodes the raw payload as a pointing report.
func (m *IncomingMessage) ParsePointing() (*PointingPayload, error) {
	var p PointingPayload
	if err := json.Unmarshal(m.Payload, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

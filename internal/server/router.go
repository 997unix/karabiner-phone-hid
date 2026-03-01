package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tonyjiang/karabiner-phone-hid/internal/hid"
	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

// ActionResolver resolves named actions to key steps.
type ActionResolver interface {
	Resolve(name string) ([]protocol.KeyStep, bool)
}

// Router decodes incoming messages, dispatches HID events, and returns responses.
type Router struct {
	dispatcher *hid.Dispatcher
	resolver   ActionResolver
}

// NewRouter creates a Router.
func NewRouter(dispatcher *hid.Dispatcher, resolver ActionResolver) *Router {
	return &Router{
		dispatcher: dispatcher,
		resolver:   resolver,
	}
}

// Route processes a raw JSON message and returns the JSON response.
func (r *Router) Route(data []byte) []byte {
	var msg protocol.IncomingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return r.encodeAck("unknown", fmt.Sprintf("invalid JSON: %v", err))
	}

	switch msg.Type {
	case "action":
		return r.handleAction(&msg)
	default:
		return r.encodeAck(msg.ID, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

func (r *Router) handleAction(msg *protocol.IncomingMessage) []byte {
	var steps []protocol.KeyStep

	switch msg.Action {
	case "keypress":
		kp, err := msg.ParseKeypress()
		if err != nil {
			return r.encodeAck(msg.ID, fmt.Sprintf("invalid keypress payload: %v", err))
		}
		steps = []protocol.KeyStep{
			{Key: kp.Key, Modifiers: kp.Modifiers},
		}

	case "sequence":
		seq, err := msg.ParseSequence()
		if err != nil {
			return r.encodeAck(msg.ID, fmt.Sprintf("invalid sequence payload: %v", err))
		}
		steps = seq.Steps

	default:
		// Named action — resolve via registry
		resolved, ok := r.resolver.Resolve(msg.Action)
		if !ok {
			return r.encodeAck(msg.ID, fmt.Sprintf("unknown action: %s", msg.Action))
		}
		steps = resolved
	}

	if err := r.dispatcher.Dispatch(steps); err != nil {
		log.Printf("[Router] dispatch error: %v", err)
		return r.encodeAck(msg.ID, fmt.Sprintf("dispatch error: %v", err))
	}

	return r.encodeOK(msg.ID)
}

func (r *Router) encodeOK(id string) []byte {
	ack := protocol.NewAckOK(id)
	data, _ := json.Marshal(ack)
	return data
}

func (r *Router) encodeAck(id string, errMsg string) []byte {
	ack := protocol.NewAckError(id, errMsg)
	data, _ := json.Marshal(ack)
	return data
}

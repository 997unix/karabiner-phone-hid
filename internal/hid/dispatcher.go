package hid

import (
	"fmt"
	"time"

	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

// Dispatcher translates KeySteps into HID report calls.
type Dispatcher struct {
	poster HIDPoster
}

// NewDispatcher creates a Dispatcher with the given HIDPoster.
func NewDispatcher(poster HIDPoster) *Dispatcher {
	return &Dispatcher{poster: poster}
}

// Dispatch sends HID reports for a sequence of key steps.
// Each step becomes: PostKeyboard(modifiers, [key]) then ReleaseAll().
// Delays between steps use the previous step's delay_ms.
func (d *Dispatcher) Dispatch(steps []protocol.KeyStep) error {
	for i, step := range steps {
		if i > 0 && steps[i-1].DelayMs > 0 {
			time.Sleep(time.Duration(steps[i-1].DelayMs) * time.Millisecond)
		}

		keyCode, ok := LookupKeyCode(step.Key)
		if !ok {
			return fmt.Errorf("unknown key: %s", step.Key)
		}

		modByte := BuildModifierByte(step.Modifiers)

		if err := d.poster.PostKeyboard(modByte, []uint16{keyCode}); err != nil {
			return fmt.Errorf("PostKeyboard: %w", err)
		}
		if err := d.poster.ReleaseAll(); err != nil {
			return fmt.Errorf("ReleaseAll: %w", err)
		}
	}
	return nil
}

// DispatchPointing sends a pointing HID report.
func (d *Dispatcher) DispatchPointing(p *protocol.PointingPayload) error {
	return d.poster.PostPointing(p.Buttons, p.X, p.Y, p.VerticalWheel, p.HorizontalWheel)
}

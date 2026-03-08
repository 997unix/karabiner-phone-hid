package hid

import "sync"

// HIDPoster sends HID keyboard and pointing device reports.
type HIDPoster interface {
	PostKeyboard(modifiers uint8, keys []uint16) error
	ReleaseAll() error
	PostPointing(buttons uint32, x, y, verticalWheel, horizontalWheel int8) error
	ReleasePointing() error
}

// MockPoster records HID calls for testing.
type MockPoster struct {
	mu    sync.Mutex
	Calls []MockCall
}

// MockCall records a single HID call.
type MockCall struct {
	Modifiers uint8
	Keys      []uint16
	Release   bool
	Pointing  bool
	Buttons   uint32
	X         int8
	Y         int8
	VWheel    int8
	HWheel    int8
}

func (m *MockPoster) PostKeyboard(modifiers uint8, keys []uint16) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	keysCopy := make([]uint16, len(keys))
	copy(keysCopy, keys)
	m.Calls = append(m.Calls, MockCall{Modifiers: modifiers, Keys: keysCopy})
	return nil
}

func (m *MockPoster) ReleaseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Release: true})
	return nil
}

func (m *MockPoster) PostPointing(buttons uint32, x, y, verticalWheel, horizontalWheel int8) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{
		Pointing: true,
		Buttons:  buttons,
		X:        x,
		Y:        y,
		VWheel:   verticalWheel,
		HWheel:   horizontalWheel,
	})
	return nil
}

func (m *MockPoster) ReleasePointing() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Pointing: true, Release: true})
	return nil
}

// Reset clears recorded calls.
func (m *MockPoster) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
}

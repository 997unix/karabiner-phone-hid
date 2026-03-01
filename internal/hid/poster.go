package hid

import "sync"

// HIDPoster sends HID keyboard reports.
type HIDPoster interface {
	PostKeyboard(modifiers uint8, keys []uint16) error
	ReleaseAll() error
}

// MockPoster records HID calls for testing.
type MockPoster struct {
	mu    sync.Mutex
	Calls []MockCall
}

// MockCall records a single PostKeyboard or ReleaseAll call.
type MockCall struct {
	Modifiers uint8
	Keys      []uint16
	Release   bool
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

// Reset clears recorded calls.
func (m *MockPoster) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
}

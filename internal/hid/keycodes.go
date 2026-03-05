package hid

// USB HID Usage IDs for keyboard keys (HID Usage Table, page 0x07).
// These are the codes sent in HID keyboard reports to Karabiner.
var keyCodes = map[string]uint16{
	// Letters
	"a": 0x04, "b": 0x05, "c": 0x06, "d": 0x07, "e": 0x08,
	"f": 0x09, "g": 0x0A, "h": 0x0B, "i": 0x0C, "j": 0x0D,
	"k": 0x0E, "l": 0x0F, "m": 0x10, "n": 0x11, "o": 0x12,
	"p": 0x13, "q": 0x14, "r": 0x15, "s": 0x16, "t": 0x17,
	"u": 0x18, "v": 0x19, "w": 0x1A, "x": 0x1B, "y": 0x1C,
	"z": 0x1D,

	// Numbers
	"1": 0x1E, "2": 0x1F, "3": 0x20, "4": 0x21, "5": 0x22,
	"6": 0x23, "7": 0x24, "8": 0x25, "9": 0x26, "0": 0x27,

	// Special keys
	"return_or_enter":          0x28,
	"escape":                   0x29,
	"delete_or_backspace":      0x2A,
	"tab":                      0x2B,
	"spacebar":                 0x2C,
	"hyphen":                   0x2D,
	"equal_sign":               0x2E,
	"open_bracket":             0x2F,
	"close_bracket":            0x30,
	"backslash":                0x31,
	"semicolon":                0x33,
	"quote":                    0x34,
	"grave_accent_and_tilde":   0x35,
	"comma":                    0x36,
	"period":                   0x37,
	"slash":                    0x38,
	"caps_lock":                0x39,

	// Function keys
	"f1": 0x3A, "f2": 0x3B, "f3": 0x3C, "f4": 0x3D,
	"f5": 0x3E, "f6": 0x3F, "f7": 0x40, "f8": 0x41,
	"f9": 0x42, "f10": 0x43, "f11": 0x44, "f12": 0x45,

	// Navigation
	"insert":         0x49,
	"right_arrow":    0x4F,
	"left_arrow":     0x50,
	"down_arrow":     0x51,
	"up_arrow":       0x52,
	"page_up":        0x4B,
	"page_down":      0x4E,
	"home":           0x4A,
	"end":            0x4D,
	"delete_forward": 0x4C,

	// Numpad
	"num_lock":          0x53,
	"numpad_divide":     0x54,
	"numpad_multiply":   0x55,
	"numpad_subtract":   0x56,
	"numpad_add":        0x57,
	"numpad_enter":      0x58,
	"numpad_1":          0x59,
	"numpad_2":          0x5A,
	"numpad_3":          0x5B,
	"numpad_4":          0x5C,
	"numpad_5":          0x5D,
	"numpad_6":          0x5E,
	"numpad_7":          0x5F,
	"numpad_8":          0x60,
	"numpad_9":          0x61,
	"numpad_0":          0x62,
	"numpad_decimal":    0x63,
}

// HID modifier bit flags (byte 0 of keyboard report).
var modifierFlags = map[string]uint8{
	"control": 0x01, // Left Control
	"shift":   0x02, // Left Shift
	"option":  0x04, // Left Alt/Option
	"command": 0x08, // Left GUI/Command
}

// LookupKeyCode returns the USB HID usage ID for a key name, or 0 and false.
func LookupKeyCode(name string) (uint16, bool) {
	code, ok := keyCodes[name]
	return code, ok
}

// LookupModifier returns the HID modifier bit flag, or 0 and false.
func LookupModifier(name string) (uint8, bool) {
	flag, ok := modifierFlags[name]
	return flag, ok
}

// BuildModifierByte combines modifier names into a single byte.
func BuildModifierByte(modifiers []string) uint8 {
	var result uint8
	for _, mod := range modifiers {
		if flag, ok := modifierFlags[mod]; ok {
			result |= flag
		}
	}
	return result
}

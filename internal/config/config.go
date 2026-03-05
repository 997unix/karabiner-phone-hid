package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

// ActionDefinition defines a named action with its key steps.
type ActionDefinition struct {
	Name  string             `json:"name"`
	Label string             `json:"label"`
	Steps []protocol.KeyStep `json:"steps"`
}

// ActionsConfig is the JSON config file structure.
type ActionsConfig struct {
	Actions []ActionDefinition `json:"actions"`
}

// ConfigDir returns the config directory path.
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "karabiner-phone-hid")
}

// ConfigFile returns the config file path.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "actions.json")
}

// LoadActions reads the config file. Returns nil if not found.
func LoadActions() (*ActionsConfig, error) {
	data, err := os.ReadFile(ConfigFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg ActionsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Registry resolves named actions to key steps.
type Registry struct {
	actions map[string]ActionDefinition
}

// NewRegistry creates a Registry with defaults, optionally merged with user config.
func NewRegistry(cfg *ActionsConfig) *Registry {
	r := &Registry{actions: defaultActions()}
	if cfg != nil {
		for _, a := range cfg.Actions {
			r.actions[a.Name] = a
		}
	}
	return r
}

// Resolve returns the key steps for a named action.
func (r *Registry) Resolve(name string) ([]protocol.KeyStep, bool) {
	a, ok := r.actions[name]
	if !ok {
		return nil, false
	}
	return a.Steps, true
}

// AllActions returns ActionInfo for all registered actions.
func (r *Registry) AllActions() []protocol.ActionInfo {
	result := make([]protocol.ActionInfo, 0, len(r.actions))
	for _, a := range r.actions {
		result = append(result, protocol.ActionInfo{Name: a.Name, Label: a.Label})
	}
	return result
}

func defaultActions() map[string]ActionDefinition {
	defaults := []ActionDefinition{
		{
			Name:  "superwhisper_toggle",
			Label: "SuperWhisper",
			Steps: []protocol.KeyStep{{Key: "spacebar", Modifiers: []string{"option"}}},
		},
		{
			Name:  "superwhisper_paste",
			Label: "SW Paste",
			Steps: []protocol.KeyStep{
				{Key: "spacebar", Modifiers: []string{"option"}, DelayMs: 500},
				{Key: "v", Modifiers: []string{"command"}},
			},
		},
		{
			Name:  "return",
			Label: "Return",
			Steps: []protocol.KeyStep{{Key: "return_or_enter", Modifiers: []string{}}},
		},
		{
			Name:  "tmux_copy_mode",
			Label: "Copy Mode",
			Steps: []protocol.KeyStep{
				{Key: "backslash", Modifiers: []string{"control"}, DelayMs: 100},
				{Key: "open_bracket", Modifiers: []string{}},
			},
		},
		{
			Name:  "tmux_scroll_up",
			Label: "Scroll Up",
			Steps: []protocol.KeyStep{{Key: "u", Modifiers: []string{"control"}}},
		},
		{
			Name:  "tmux_scroll_down",
			Label: "Scroll Down",
			Steps: []protocol.KeyStep{{Key: "d", Modifiers: []string{"control"}}},
		},
		{
			Name:  "tmux_page_up",
			Label: "Page Up",
			Steps: []protocol.KeyStep{{Key: "b", Modifiers: []string{"control"}}},
		},
		{
			Name:  "tmux_page_down",
			Label: "Page Down",
			Steps: []protocol.KeyStep{{Key: "f", Modifiers: []string{"control"}}},
		},
		{
			Name:  "tmux_exit_copy",
			Label: "Exit Copy",
			Steps: []protocol.KeyStep{{Key: "q", Modifiers: []string{}}},
		},
		{
			Name:  "tmux_prefix",
			Label: "Prefix",
			Steps: []protocol.KeyStep{{Key: "backslash", Modifiers: []string{"control"}}},
		},
		// Tmux extras
		{
			Name:  "tmux_split_h",
			Label: "Split H",
			Steps: []protocol.KeyStep{
				{Key: "backslash", Modifiers: []string{"control"}, DelayMs: 100},
				{Key: "quote", Modifiers: []string{"shift"}},
			},
		},
		{
			Name:  "tmux_split_v",
			Label: "Split V",
			Steps: []protocol.KeyStep{
				{Key: "backslash", Modifiers: []string{"control"}, DelayMs: 100},
				{Key: "5", Modifiers: []string{"shift"}},
			},
		},
		{
			Name:  "tmux_next_pane",
			Label: "Next Pane",
			Steps: []protocol.KeyStep{
				{Key: "backslash", Modifiers: []string{"control"}, DelayMs: 100},
				{Key: "o", Modifiers: []string{}},
			},
		},
		{
			Name:  "tmux_next_window",
			Label: "Next Win",
			Steps: []protocol.KeyStep{
				{Key: "backslash", Modifiers: []string{"control"}, DelayMs: 100},
				{Key: "n", Modifiers: []string{}},
			},
		},
		// Teammate (Claude Code / Copilot inline suggestions)
		{
			Name:  "teammate_accept",
			Label: "Accept",
			Steps: []protocol.KeyStep{{Key: "tab", Modifiers: []string{}}},
		},
		{
			Name:  "teammate_reject",
			Label: "Reject",
			Steps: []protocol.KeyStep{{Key: "escape", Modifiers: []string{}}},
		},
		{
			Name:  "teammate_attention",
			Label: "tmate-attn",
			Steps: []protocol.KeyStep{{Key: "c", Modifiers: []string{"control"}}},
		},
		{
			Name:  "paste",
			Label: "Paste",
			Steps: []protocol.KeyStep{{Key: "v", Modifiers: []string{"command"}}},
		},
		// Media (F-keys = media controls on macOS by default)
		{
			Name:  "media_prev",
			Label: "Prev",
			Steps: []protocol.KeyStep{{Key: "f7", Modifiers: []string{}}},
		},
		{
			Name:  "media_play_pause",
			Label: "Play/Pause",
			Steps: []protocol.KeyStep{{Key: "f8", Modifiers: []string{}}},
		},
		{
			Name:  "media_next",
			Label: "Next",
			Steps: []protocol.KeyStep{{Key: "f9", Modifiers: []string{}}},
		},
		{
			Name:  "media_mute",
			Label: "Mute",
			Steps: []protocol.KeyStep{{Key: "f10", Modifiers: []string{}}},
		},
		{
			Name:  "media_vol_down",
			Label: "Vol Down",
			Steps: []protocol.KeyStep{{Key: "f11", Modifiers: []string{}}},
		},
		{
			Name:  "media_vol_up",
			Label: "Vol Up",
			Steps: []protocol.KeyStep{{Key: "f12", Modifiers: []string{}}},
		},
		// YouTube Remote
		{
			Name:  "yt_play_pause",
			Label: "Play/Pause",
			Steps: []protocol.KeyStep{{Key: "k", Modifiers: []string{}}},
		},
		{
			Name:  "yt_back_10",
			Label: "Back 10s",
			Steps: []protocol.KeyStep{{Key: "j", Modifiers: []string{}}},
		},
		{
			Name:  "yt_fwd_10",
			Label: "Fwd 10s",
			Steps: []protocol.KeyStep{{Key: "l", Modifiers: []string{}}},
		},
		{
			Name:  "yt_mute",
			Label: "Mute",
			Steps: []protocol.KeyStep{{Key: "m", Modifiers: []string{}}},
		},
		{
			Name:  "yt_fullscreen",
			Label: "Fullscreen",
			Steps: []protocol.KeyStep{{Key: "f", Modifiers: []string{}}},
		},
		{
			Name:  "yt_captions",
			Label: "Captions",
			Steps: []protocol.KeyStep{{Key: "c", Modifiers: []string{}}},
		},
		{
			Name:  "yt_speed_up",
			Label: "Speed Up",
			Steps: []protocol.KeyStep{{Key: "period", Modifiers: []string{"shift"}}},
		},
		{
			Name:  "yt_speed_down",
			Label: "Speed Down",
			Steps: []protocol.KeyStep{{Key: "comma", Modifiers: []string{"shift"}}},
		},
		{
			Name:  "yt_next_video",
			Label: "Next Video",
			Steps: []protocol.KeyStep{{Key: "n", Modifiers: []string{"shift"}}},
		},
		{
			Name:  "yt_prev_video",
			Label: "Prev Video",
			Steps: []protocol.KeyStep{{Key: "p", Modifiers: []string{"shift"}}},
		},
		{
			Name:  "yt_vol_up",
			Label: "Vol Up",
			Steps: []protocol.KeyStep{{Key: "up_arrow", Modifiers: []string{}}},
		},
		{
			Name:  "yt_vol_down",
			Label: "Vol Down",
			Steps: []protocol.KeyStep{{Key: "down_arrow", Modifiers: []string{}}},
		},
		// Keyboard keys
		{
			Name:  "key_escape",
			Label: "Esc",
			Steps: []protocol.KeyStep{{Key: "escape", Modifiers: []string{}}},
		},
		{
			Name:  "key_tab",
			Label: "Tab",
			Steps: []protocol.KeyStep{{Key: "tab", Modifiers: []string{}}},
		},
		{
			Name:  "key_backspace",
			Label: "Bksp",
			Steps: []protocol.KeyStep{{Key: "delete_or_backspace", Modifiers: []string{}}},
		},
		{
			Name:  "key_delete",
			Label: "Del",
			Steps: []protocol.KeyStep{{Key: "delete_forward", Modifiers: []string{}}},
		},
		{
			Name:  "key_space",
			Label: "Space",
			Steps: []protocol.KeyStep{{Key: "spacebar", Modifiers: []string{}}},
		},
		{
			Name:  "key_enter",
			Label: "Enter",
			Steps: []protocol.KeyStep{{Key: "return_or_enter", Modifiers: []string{}}},
		},
		{
			Name:  "mod_ctrl",
			Label: "Ctrl",
			Steps: []protocol.KeyStep{},
		},
		{
			Name:  "mod_shift",
			Label: "Shift",
			Steps: []protocol.KeyStep{},
		},
		{
			Name:  "mod_option",
			Label: "Opt",
			Steps: []protocol.KeyStep{},
		},
		{
			Name:  "mod_command",
			Label: "Cmd",
			Steps: []protocol.KeyStep{},
		},
		// Symbols
		{
			Name:  "key_hyphen",
			Label: "-",
			Steps: []protocol.KeyStep{{Key: "hyphen", Modifiers: []string{}}},
		},
		{
			Name:  "key_equal",
			Label: "=",
			Steps: []protocol.KeyStep{{Key: "equal_sign", Modifiers: []string{}}},
		},
		{
			Name:  "key_open_bracket",
			Label: "[",
			Steps: []protocol.KeyStep{{Key: "open_bracket", Modifiers: []string{}}},
		},
		{
			Name:  "key_close_bracket",
			Label: "]",
			Steps: []protocol.KeyStep{{Key: "close_bracket", Modifiers: []string{}}},
		},
		{
			Name:  "key_backslash",
			Label: "\\",
			Steps: []protocol.KeyStep{{Key: "backslash", Modifiers: []string{}}},
		},
		{
			Name:  "key_semicolon",
			Label: ";",
			Steps: []protocol.KeyStep{{Key: "semicolon", Modifiers: []string{}}},
		},
		{
			Name:  "key_quote",
			Label: "'",
			Steps: []protocol.KeyStep{{Key: "quote", Modifiers: []string{}}},
		},
		{
			Name:  "key_grave",
			Label: "`",
			Steps: []protocol.KeyStep{{Key: "grave_accent_and_tilde", Modifiers: []string{}}},
		},
		{
			Name:  "key_comma",
			Label: ",",
			Steps: []protocol.KeyStep{{Key: "comma", Modifiers: []string{}}},
		},
		{
			Name:  "key_period",
			Label: ".",
			Steps: []protocol.KeyStep{{Key: "period", Modifiers: []string{}}},
		},
		{
			Name:  "key_slash",
			Label: "/",
			Steps: []protocol.KeyStep{{Key: "slash", Modifiers: []string{}}},
		},
		// Function keys
		{
			Name:  "key_f1",
			Label: "F1",
			Steps: []protocol.KeyStep{{Key: "f1", Modifiers: []string{}}},
		},
		{
			Name:  "key_f2",
			Label: "F2",
			Steps: []protocol.KeyStep{{Key: "f2", Modifiers: []string{}}},
		},
		{
			Name:  "key_f3",
			Label: "F3",
			Steps: []protocol.KeyStep{{Key: "f3", Modifiers: []string{}}},
		},
		{
			Name:  "key_f4",
			Label: "F4",
			Steps: []protocol.KeyStep{{Key: "f4", Modifiers: []string{}}},
		},
		{
			Name:  "key_f5",
			Label: "F5",
			Steps: []protocol.KeyStep{{Key: "f5", Modifiers: []string{}}},
		},
		{
			Name:  "key_f6",
			Label: "F6",
			Steps: []protocol.KeyStep{{Key: "f6", Modifiers: []string{}}},
		},
		{
			Name:  "key_f7",
			Label: "F7",
			Steps: []protocol.KeyStep{{Key: "f7", Modifiers: []string{}}},
		},
		{
			Name:  "key_f8",
			Label: "F8",
			Steps: []protocol.KeyStep{{Key: "f8", Modifiers: []string{}}},
		},
		{
			Name:  "key_f9",
			Label: "F9",
			Steps: []protocol.KeyStep{{Key: "f9", Modifiers: []string{}}},
		},
		{
			Name:  "key_f10",
			Label: "F10",
			Steps: []protocol.KeyStep{{Key: "f10", Modifiers: []string{}}},
		},
		{
			Name:  "key_f11",
			Label: "F11",
			Steps: []protocol.KeyStep{{Key: "f11", Modifiers: []string{}}},
		},
		{
			Name:  "key_f12",
			Label: "F12",
			Steps: []protocol.KeyStep{{Key: "f12", Modifiers: []string{}}},
		},
		// Letter keys
		{Name: "key_a", Label: "A", Steps: []protocol.KeyStep{{Key: "a", Modifiers: []string{}}}},
		{Name: "key_b", Label: "B", Steps: []protocol.KeyStep{{Key: "b", Modifiers: []string{}}}},
		{Name: "key_c", Label: "C", Steps: []protocol.KeyStep{{Key: "c", Modifiers: []string{}}}},
		{Name: "key_d", Label: "D", Steps: []protocol.KeyStep{{Key: "d", Modifiers: []string{}}}},
		{Name: "key_e", Label: "E", Steps: []protocol.KeyStep{{Key: "e", Modifiers: []string{}}}},
		{Name: "key_f", Label: "F", Steps: []protocol.KeyStep{{Key: "f", Modifiers: []string{}}}},
		{Name: "key_g", Label: "G", Steps: []protocol.KeyStep{{Key: "g", Modifiers: []string{}}}},
		{Name: "key_h", Label: "H", Steps: []protocol.KeyStep{{Key: "h", Modifiers: []string{}}}},
		{Name: "key_i", Label: "I", Steps: []protocol.KeyStep{{Key: "i", Modifiers: []string{}}}},
		{Name: "key_j", Label: "J", Steps: []protocol.KeyStep{{Key: "j", Modifiers: []string{}}}},
		{Name: "key_k", Label: "K", Steps: []protocol.KeyStep{{Key: "k", Modifiers: []string{}}}},
		{Name: "key_l", Label: "L", Steps: []protocol.KeyStep{{Key: "l", Modifiers: []string{}}}},
		{Name: "key_m", Label: "M", Steps: []protocol.KeyStep{{Key: "m", Modifiers: []string{}}}},
		{Name: "key_n", Label: "N", Steps: []protocol.KeyStep{{Key: "n", Modifiers: []string{}}}},
		{Name: "key_o", Label: "O", Steps: []protocol.KeyStep{{Key: "o", Modifiers: []string{}}}},
		{Name: "key_p", Label: "P", Steps: []protocol.KeyStep{{Key: "p", Modifiers: []string{}}}},
		{Name: "key_q", Label: "Q", Steps: []protocol.KeyStep{{Key: "q", Modifiers: []string{}}}},
		{Name: "key_r", Label: "R", Steps: []protocol.KeyStep{{Key: "r", Modifiers: []string{}}}},
		{Name: "key_s", Label: "S", Steps: []protocol.KeyStep{{Key: "s", Modifiers: []string{}}}},
		{Name: "key_t", Label: "T", Steps: []protocol.KeyStep{{Key: "t", Modifiers: []string{}}}},
		{Name: "key_u", Label: "U", Steps: []protocol.KeyStep{{Key: "u", Modifiers: []string{}}}},
		{Name: "key_v", Label: "V", Steps: []protocol.KeyStep{{Key: "v", Modifiers: []string{}}}},
		{Name: "key_w", Label: "W", Steps: []protocol.KeyStep{{Key: "w", Modifiers: []string{}}}},
		{Name: "key_x", Label: "X", Steps: []protocol.KeyStep{{Key: "x", Modifiers: []string{}}}},
		{Name: "key_y", Label: "Y", Steps: []protocol.KeyStep{{Key: "y", Modifiers: []string{}}}},
		{Name: "key_z", Label: "Z", Steps: []protocol.KeyStep{{Key: "z", Modifiers: []string{}}}},
		// Number keys
		{Name: "key_1", Label: "1", Steps: []protocol.KeyStep{{Key: "1", Modifiers: []string{}}}},
		{Name: "key_2", Label: "2", Steps: []protocol.KeyStep{{Key: "2", Modifiers: []string{}}}},
		{Name: "key_3", Label: "3", Steps: []protocol.KeyStep{{Key: "3", Modifiers: []string{}}}},
		{Name: "key_4", Label: "4", Steps: []protocol.KeyStep{{Key: "4", Modifiers: []string{}}}},
		{Name: "key_5", Label: "5", Steps: []protocol.KeyStep{{Key: "5", Modifiers: []string{}}}},
		{Name: "key_6", Label: "6", Steps: []protocol.KeyStep{{Key: "6", Modifiers: []string{}}}},
		{Name: "key_7", Label: "7", Steps: []protocol.KeyStep{{Key: "7", Modifiers: []string{}}}},
		{Name: "key_8", Label: "8", Steps: []protocol.KeyStep{{Key: "8", Modifiers: []string{}}}},
		{Name: "key_9", Label: "9", Steps: []protocol.KeyStep{{Key: "9", Modifiers: []string{}}}},
		{Name: "key_0", Label: "0", Steps: []protocol.KeyStep{{Key: "0", Modifiers: []string{}}}},
		// Navigation keys
		{
			Name:  "key_home",
			Label: "Home",
			Steps: []protocol.KeyStep{{Key: "home", Modifiers: []string{}}},
		},
		{
			Name:  "key_end",
			Label: "End",
			Steps: []protocol.KeyStep{{Key: "end", Modifiers: []string{}}},
		},
		{
			Name:  "key_page_up",
			Label: "PgUp",
			Steps: []protocol.KeyStep{{Key: "page_up", Modifiers: []string{}}},
		},
		{
			Name:  "key_page_down",
			Label: "PgDn",
			Steps: []protocol.KeyStep{{Key: "page_down", Modifiers: []string{}}},
		},
		// Arrow keys
		{
			Name:  "arrow_up",
			Label: "Up",
			Steps: []protocol.KeyStep{{Key: "up_arrow", Modifiers: []string{}}},
		},
		{
			Name:  "arrow_down",
			Label: "Down",
			Steps: []protocol.KeyStep{{Key: "down_arrow", Modifiers: []string{}}},
		},
		{
			Name:  "arrow_left",
			Label: "Left",
			Steps: []protocol.KeyStep{{Key: "left_arrow", Modifiers: []string{}}},
		},
		{
			Name:  "arrow_right",
			Label: "Right",
			Steps: []protocol.KeyStep{{Key: "right_arrow", Modifiers: []string{}}},
		},
		// Insert key
		{
			Name:  "key_insert",
			Label: "Ins",
			Steps: []protocol.KeyStep{{Key: "insert", Modifiers: []string{}}},
		},
		// Numpad keys
		{Name: "numpad_0", Label: "0", Steps: []protocol.KeyStep{{Key: "numpad_0", Modifiers: []string{}}}},
		{Name: "numpad_1", Label: "1", Steps: []protocol.KeyStep{{Key: "numpad_1", Modifiers: []string{}}}},
		{Name: "numpad_2", Label: "2", Steps: []protocol.KeyStep{{Key: "numpad_2", Modifiers: []string{}}}},
		{Name: "numpad_3", Label: "3", Steps: []protocol.KeyStep{{Key: "numpad_3", Modifiers: []string{}}}},
		{Name: "numpad_4", Label: "4", Steps: []protocol.KeyStep{{Key: "numpad_4", Modifiers: []string{}}}},
		{Name: "numpad_5", Label: "5", Steps: []protocol.KeyStep{{Key: "numpad_5", Modifiers: []string{}}}},
		{Name: "numpad_6", Label: "6", Steps: []protocol.KeyStep{{Key: "numpad_6", Modifiers: []string{}}}},
		{Name: "numpad_7", Label: "7", Steps: []protocol.KeyStep{{Key: "numpad_7", Modifiers: []string{}}}},
		{Name: "numpad_8", Label: "8", Steps: []protocol.KeyStep{{Key: "numpad_8", Modifiers: []string{}}}},
		{Name: "numpad_9", Label: "9", Steps: []protocol.KeyStep{{Key: "numpad_9", Modifiers: []string{}}}},
		{Name: "numpad_divide", Label: "/", Steps: []protocol.KeyStep{{Key: "numpad_divide", Modifiers: []string{}}}},
		{Name: "numpad_multiply", Label: "*", Steps: []protocol.KeyStep{{Key: "numpad_multiply", Modifiers: []string{}}}},
		{Name: "numpad_subtract", Label: "-", Steps: []protocol.KeyStep{{Key: "numpad_subtract", Modifiers: []string{}}}},
		{Name: "numpad_add", Label: "+", Steps: []protocol.KeyStep{{Key: "numpad_add", Modifiers: []string{}}}},
		{Name: "numpad_enter", Label: "Enter", Steps: []protocol.KeyStep{{Key: "numpad_enter", Modifiers: []string{}}}},
		{Name: "numpad_decimal", Label: ".", Steps: []protocol.KeyStep{{Key: "numpad_decimal", Modifiers: []string{}}}},
	}

	m := make(map[string]ActionDefinition, len(defaults))
	for _, a := range defaults {
		m[a.Name] = a
	}
	return m
}

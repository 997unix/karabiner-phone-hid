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
		// Zoom shortcuts
		{
			Name:  "zoom_mute",
			Label: "Mute",
			Steps: []protocol.KeyStep{{Key: "a", Modifiers: []string{"command", "shift"}}},
		},
		{
			Name:  "zoom_video",
			Label: "Video",
			Steps: []protocol.KeyStep{{Key: "v", Modifiers: []string{"command", "shift"}}},
		},
		{
			Name:  "zoom_share",
			Label: "Share",
			Steps: []protocol.KeyStep{{Key: "s", Modifiers: []string{"command", "shift"}}},
		},
		{
			Name:  "zoom_chat",
			Label: "Chat",
			Steps: []protocol.KeyStep{{Key: "h", Modifiers: []string{"command", "shift"}}},
		},
		{
			Name:  "zoom_hand",
			Label: "Hand",
			Steps: []protocol.KeyStep{{Key: "y", Modifiers: []string{"option"}}},
		},
		{
			Name:  "zoom_leave",
			Label: "Leave",
			Steps: []protocol.KeyStep{{Key: "w", Modifiers: []string{"command"}}},
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
	}

	m := make(map[string]ActionDefinition, len(defaults))
	for _, a := range defaults {
		m[a.Name] = a
	}
	return m
}

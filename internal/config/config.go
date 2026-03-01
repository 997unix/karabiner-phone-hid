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
	}

	m := make(map[string]ActionDefinition, len(defaults))
	for _, a := range defaults {
		m[a.Name] = a
	}
	return m
}

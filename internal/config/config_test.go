package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

func TestDefaultRegistryResolvesKnownActions(t *testing.T) {
	r := NewRegistry(nil)

	tests := []struct {
		name    string
		wantKey string
	}{
		{"superwhisper_toggle", "spacebar"},
		{"return", "return_or_enter"},
		{"tmux_prefix", "backslash"},
		{"tmux_copy_mode", "backslash"},
	}

	for _, tt := range tests {
		steps, ok := r.Resolve(tt.name)
		if !ok {
			t.Errorf("Resolve(%q) not found", tt.name)
			continue
		}
		if len(steps) == 0 {
			t.Errorf("Resolve(%q) returned empty steps", tt.name)
			continue
		}
		if steps[0].Key != tt.wantKey {
			t.Errorf("Resolve(%q)[0].Key = %q, want %q", tt.name, steps[0].Key, tt.wantKey)
		}
	}
}

func TestDefaultRegistryUnknownAction(t *testing.T) {
	r := NewRegistry(nil)
	_, ok := r.Resolve("nonexistent")
	if ok {
		t.Error("Resolve(\"nonexistent\") should return false")
	}
}

func TestRegistryMergesUserConfig(t *testing.T) {
	cfg := &ActionsConfig{
		Actions: []ActionDefinition{
			{
				Name:  "custom_action",
				Label: "Custom",
				Steps: []protocol.KeyStep{{Key: "x", Modifiers: []string{"command"}}},
			},
		},
	}

	r := NewRegistry(cfg)

	// Custom action should exist
	steps, ok := r.Resolve("custom_action")
	if !ok {
		t.Fatal("custom_action not found")
	}
	if steps[0].Key != "x" {
		t.Errorf("custom_action key = %q, want x", steps[0].Key)
	}

	// Default actions should still exist
	_, ok = r.Resolve("superwhisper_toggle")
	if !ok {
		t.Error("default action superwhisper_toggle lost after merge")
	}
}

func TestRegistryOverridesDefault(t *testing.T) {
	cfg := &ActionsConfig{
		Actions: []ActionDefinition{
			{
				Name:  "return",
				Label: "Custom Return",
				Steps: []protocol.KeyStep{{Key: "a", Modifiers: []string{}}},
			},
		},
	}

	r := NewRegistry(cfg)
	steps, ok := r.Resolve("return")
	if !ok {
		t.Fatal("return not found")
	}
	if steps[0].Key != "a" {
		t.Errorf("overridden return key = %q, want a", steps[0].Key)
	}
}

func TestAllActions(t *testing.T) {
	r := NewRegistry(nil)
	actions := r.AllActions()

	if len(actions) != 29 {
		t.Errorf("AllActions len = %d, want 29", len(actions))
	}

	// Check that each has name and label
	for _, a := range actions {
		if a.Name == "" || a.Label == "" {
			t.Errorf("action with empty name or label: %+v", a)
		}
	}
}

func TestTeammateActions(t *testing.T) {
	r := NewRegistry(nil)

	tests := []struct {
		name     string
		wantKey  string
		wantMods []string
	}{
		{"teammate_accept", "tab", nil},
		{"teammate_reject", "escape", nil},
		{"teammate_interrupt", "c", []string{"control"}},
	}

	for _, tt := range tests {
		steps, ok := r.Resolve(tt.name)
		if !ok {
			t.Errorf("Resolve(%q) not found", tt.name)
			continue
		}
		if steps[0].Key != tt.wantKey {
			t.Errorf("Resolve(%q)[0].Key = %q, want %q", tt.name, steps[0].Key, tt.wantKey)
		}
		if tt.wantMods != nil && steps[0].Modifiers[0] != tt.wantMods[0] {
			t.Errorf("Resolve(%q)[0].Modifiers = %v, want %v", tt.name, steps[0].Modifiers, tt.wantMods)
		}
	}
}

func TestMediaActions(t *testing.T) {
	r := NewRegistry(nil)

	tests := []struct {
		name    string
		wantKey string
	}{
		{"media_play_pause", "f8"},
		{"media_next", "f9"},
		{"media_prev", "f7"},
		{"media_vol_up", "f12"},
		{"media_vol_down", "f11"},
		{"media_mute", "f10"},
	}

	for _, tt := range tests {
		steps, ok := r.Resolve(tt.name)
		if !ok {
			t.Errorf("Resolve(%q) not found", tt.name)
			continue
		}
		if steps[0].Key != tt.wantKey {
			t.Errorf("Resolve(%q)[0].Key = %q, want %q", tt.name, steps[0].Key, tt.wantKey)
		}
	}
}

func TestZoomActions(t *testing.T) {
	r := NewRegistry(nil)

	tests := []struct {
		name     string
		wantKey  string
		wantMods []string
	}{
		{"zoom_mute", "a", []string{"command", "shift"}},
		{"zoom_video", "v", []string{"command", "shift"}},
		{"zoom_share", "s", []string{"command", "shift"}},
		{"zoom_chat", "h", []string{"command", "shift"}},
		{"zoom_hand", "y", []string{"option"}},
		{"zoom_leave", "w", []string{"command"}},
	}

	for _, tt := range tests {
		steps, ok := r.Resolve(tt.name)
		if !ok {
			t.Errorf("Resolve(%q) not found", tt.name)
			continue
		}
		if steps[0].Key != tt.wantKey {
			t.Errorf("Resolve(%q)[0].Key = %q, want %q", tt.name, steps[0].Key, tt.wantKey)
		}
		for i, mod := range tt.wantMods {
			found := false
			for _, m := range steps[0].Modifiers {
				if m == mod {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Resolve(%q)[0].Modifiers missing %q (index %d), got %v", tt.name, mod, i, steps[0].Modifiers)
			}
		}
	}
}

func TestTmuxExtraActions(t *testing.T) {
	r := NewRegistry(nil)

	tests := []struct {
		name      string
		wantSteps int
		lastKey   string
	}{
		{"tmux_split_h", 2, "quote"},
		{"tmux_split_v", 2, "5"},
		{"tmux_next_pane", 2, "o"},
		{"tmux_next_window", 2, "n"},
	}

	for _, tt := range tests {
		steps, ok := r.Resolve(tt.name)
		if !ok {
			t.Errorf("Resolve(%q) not found", tt.name)
			continue
		}
		if len(steps) != tt.wantSteps {
			t.Errorf("Resolve(%q) steps len = %d, want %d", tt.name, len(steps), tt.wantSteps)
			continue
		}
		// First step should be tmux prefix (Ctrl+\)
		if steps[0].Key != "backslash" || steps[0].Modifiers[0] != "control" {
			t.Errorf("Resolve(%q)[0] = %+v, want backslash+control", tt.name, steps[0])
		}
		if steps[1].Key != tt.lastKey {
			t.Errorf("Resolve(%q)[1].Key = %q, want %q", tt.name, steps[1].Key, tt.lastKey)
		}
	}
}

func TestSuperwhisperPasteHasDelay(t *testing.T) {
	r := NewRegistry(nil)
	steps, ok := r.Resolve("superwhisper_paste")
	if !ok {
		t.Fatal("superwhisper_paste not found")
	}
	if len(steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(steps))
	}
	if steps[0].DelayMs != 500 {
		t.Errorf("step[0].DelayMs = %d, want 500", steps[0].DelayMs)
	}
}

func TestTmuxCopyModeSequence(t *testing.T) {
	r := NewRegistry(nil)
	steps, ok := r.Resolve("tmux_copy_mode")
	if !ok {
		t.Fatal("tmux_copy_mode not found")
	}
	if len(steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(steps))
	}
	if steps[0].Key != "backslash" || steps[0].Modifiers[0] != "control" {
		t.Errorf("step[0] = %+v, want backslash+control", steps[0])
	}
	if steps[1].Key != "open_bracket" {
		t.Errorf("step[1].Key = %q, want open_bracket", steps[1].Key)
	}
}

func TestLoadActionsFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "actions.json")

	cfg := ActionsConfig{
		Actions: []ActionDefinition{
			{Name: "file_action", Label: "File Action", Steps: []protocol.KeyStep{{Key: "f", Modifiers: []string{}}}},
		},
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read it back
	readData, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var loaded ActionsConfig
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(loaded.Actions) != 1 {
		t.Fatalf("actions len = %d, want 1", len(loaded.Actions))
	}
	if loaded.Actions[0].Name != "file_action" {
		t.Errorf("name = %q, want file_action", loaded.Actions[0].Name)
	}
}

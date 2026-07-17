package domain

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// RainbowOption specifies whether the rainbow bar is enabled and optionally custom colors.
type RainbowOption struct {
	Enabled bool
	Colors  []string
}

// UnmarshalTOML implements custom TOML unmarshaling for RainbowOption.
// It can unmarshal a boolean (true/false) or an array of strings (custom colors).
func (r *RainbowOption) UnmarshalTOML(data []byte) error {
	// Try unmarshaling as boolean
	var b bool
	if err := toml.Unmarshal(data, &b); err == nil {
		r.Enabled = b
		r.Colors = nil
		return nil
	}

	// Try unmarshaling as string slice
	var s []string
	if err := toml.Unmarshal(data, &s); err == nil {
		r.Enabled = true
		r.Colors = s
		return nil
	}

	return fmt.Errorf("rainbow option must be a boolean or a list of color strings")
}

// Config describes the resolved configuration.
type Config struct {
	Editor             string        `toml:"editor"`
	Border             string        `toml:"border"`
	InteractiveDefault bool          `toml:"interactive_default"`
	ListSpacing        string        `toml:"list_spacing"`
	DefaultUnits       string        `toml:"default_units"`
	AlarmSound         string        `toml:"alarm_sound"`
	TimeRemaining      string        `toml:"time_remaining"`
	TimeStart          string        `toml:"time_start"`
	BarBg              string        `toml:"bar_bg"`
	BarFg              string        `toml:"bar_fg"`
	HelpText           string        `toml:"help_text"`
	UpdateTmuxWindow   bool          `toml:"update_tmux_window"`
	TmuxProgressBar    bool          `toml:"tmux_progress_bar"`
	Rainbow            RainbowOption `toml:"rainbow"`
	RainbowBar         RainbowOption `toml:"rainbow_bar"`
}

// DefaultConfig returns the default configuration values.
func DefaultConfig() Config {
	defaultRainbow := RainbowOption{
		Enabled: true,
		Colors:  nil,
	}
	return Config{
		Editor:             "nvim",
		Border:             "08",
		InteractiveDefault: true,
		ListSpacing:        "space",
		DefaultUnits:       "minutes",
		AlarmSound:         "",
		TimeRemaining:      "14",
		TimeStart:          "07",
		BarBg:              "08",
		BarFg:              "02",
		HelpText:           "08",
		UpdateTmuxWindow:   false,
		TmuxProgressBar:    true,
		Rainbow:            defaultRainbow,
		RainbowBar:         defaultRainbow,
	}
}

func xdgHome(envKey, fallbackSuffix string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, fallbackSuffix)
}

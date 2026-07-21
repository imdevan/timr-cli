package domain

import (
	"os"
	"path/filepath"
)

// RainbowOption specifies whether the rainbow bar is enabled and optionally custom colors.
type RainbowOption struct {
	Enabled bool
	Colors  []string
}

// Config describes the resolved configuration.
type Config struct {
	Editor             string        `toml:"editor"`
	Border             string        `toml:"border"`
	InteractiveDefault bool          `toml:"interactive_default"`
	ListSpacing        string        `toml:"list_spacing"`
	DefaultUnits       string        `toml:"default_units"`
	DefaultTimer       string        `toml:"default_timer"`
	AlarmSound         string        `toml:"alarm_sound"`
	TimeRemaining      string        `toml:"time_remaining"`
	TimeStart          string        `toml:"time_start"`
	BarBg              string        `toml:"bar_bg"`
	BarFg              []string      `toml:"bar_fg"`
	HelpText           string        `toml:"help_text"`
	UpdateTmuxWindow   bool          `toml:"update_tmux_window"`
	TmuxProgressBar    bool          `toml:"tmux_progress_bar"`
	TmuxInverted       bool          `toml:"tmux_inverted"`
	FullWidth          bool          `toml:"full_width"`
	FullTUI            bool          `toml:"full_tui"`
	Pomodoro           []int         `toml:"pomodoro"`
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
		DefaultTimer:       "",
		AlarmSound:         "",
		TimeRemaining:      "14",
		TimeStart:          "07",
		BarBg:              "08",
		BarFg:              []string{"02", "03", "01"},
		HelpText:           "08",
		UpdateTmuxWindow:   false,
		TmuxProgressBar:    true,
		TmuxInverted:       false,
		FullWidth:          true,
		FullTUI:            true,
		Pomodoro:           []int{25, 5, 25, 5, 25, 20},
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

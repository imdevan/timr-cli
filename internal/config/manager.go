package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/timr/internal/domain"
	"github.com/timr/internal/utils"
)

// ManagerImpl loads and saves configuration files.
type ManagerImpl struct {
	cwd string
}

// NewManager returns a config manager rooted at the provided cwd.
func NewManager(cwd string) *ManagerImpl {
	return &ManagerImpl{cwd: cwd}
}

// LoadWithOverride loads config from a specific path, layered on defaults.
func (m *ManagerImpl) LoadWithOverride(path string) (domain.Config, error) {
	config := domain.DefaultConfig()
	if strings.TrimSpace(path) == "" {
		return m.Load()
	}
	partial, err := readConfig(path)
	if err != nil {
		return domain.Config{}, err
	}
	if partial != nil {
		applyPartial(&config, partial)
	}
	return config, nil
}

// Load reads config with precedence: defaults < global < local.
func (m *ManagerImpl) Load() (domain.Config, error) {
	config := domain.DefaultConfig()

	globalPath := utils.ConfigPathGlobal()
	if partial, err := readConfig(globalPath); err != nil {
		return domain.Config{}, err
	} else if partial != nil {
		applyPartial(&config, partial)
	}

	localPath := utils.ConfigPathLocal(m.cwd)
	if partial, err := readConfig(localPath); err != nil {
		return domain.Config{}, err
	} else if partial != nil {
		applyPartial(&config, partial)
	}

	return config, nil
}

// Save persists config to the global config path.
func (m *ManagerImpl) Save(config domain.Config) error {
	path := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Exists reports whether a local or global config file exists.
func (m *ManagerImpl) Exists() (bool, error) {
	globalPath := utils.ConfigPathGlobal()
	if exists, err := fileExists(globalPath); err != nil {
		return false, err
	} else if exists {
		return true, nil
	}
	localPath := utils.ConfigPathLocal(m.cwd)
	return fileExists(localPath)
}

type partialConfig struct {
	Editor               *string `toml:"editor"`
	Border               *string `toml:"border"`
	InteractiveDefault   *bool   `toml:"interactive_default"`
	ListSpacing          *string `toml:"list_spacing"`
	DefaultUnits         *string `toml:"default_units"`
	AlarmSound           *string `toml:"alarm_sound"`
	TimeRemaining        *string `toml:"time_remaining"`
	TimeStart            *string `toml:"time_start"`
	BarBg                *string `toml:"bar_bg"`
	BarFg                any     `toml:"bar_fg"`
	HelpText             *string `toml:"help_text"`
	UpdateTmuxWindow     *bool `toml:"update_tmux_window"`
	TmuxProgressBar      *bool `toml:"tmux_progress_bar"`
	TmuxInverted         *bool `toml:"tmux_inverted"`
	FullWidth            *bool `toml:"full_width"`
	FullTUI              *bool `toml:"full_tui"`
	Rainbow              any   `toml:"rainbow"`
	RainbowBar           any   `toml:"rainbow_bar"`
}

func readConfig(path string) (*partialConfig, error) {
	if exists, err := fileExists(path); err != nil || !exists {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var partial partialConfig
	if err := toml.Unmarshal(data, &partial); err != nil {
		return nil, err
	}
	return &partial, nil
}

func applyPartial(config *domain.Config, partial *partialConfig) {
	if partial.Editor != nil {
		config.Editor = *partial.Editor
	}
	if partial.Border != nil {
		config.Border = *partial.Border
	}
	if partial.InteractiveDefault != nil {
		config.InteractiveDefault = *partial.InteractiveDefault
	}
	if partial.ListSpacing != nil {
		config.ListSpacing = *partial.ListSpacing
	}
	if partial.DefaultUnits != nil {
		config.DefaultUnits = *partial.DefaultUnits
	}
	if partial.AlarmSound != nil {
		config.AlarmSound = *partial.AlarmSound
	}
	if partial.TimeRemaining != nil {
		config.TimeRemaining = *partial.TimeRemaining
	}
	if partial.TimeStart != nil {
		config.TimeStart = *partial.TimeStart
	}
	if partial.BarBg != nil {
		config.BarBg = *partial.BarBg
	}
	if partial.BarFg != nil {
		config.BarFg = parseBarFg(partial.BarFg)
	}
	if partial.HelpText != nil {
		config.HelpText = *partial.HelpText
	}
	if partial.UpdateTmuxWindow != nil {
		config.UpdateTmuxWindow = *partial.UpdateTmuxWindow
	}
	if partial.TmuxProgressBar != nil {
		config.TmuxProgressBar = *partial.TmuxProgressBar
	}
	if partial.TmuxInverted != nil {
		config.TmuxInverted = *partial.TmuxInverted
	}
	if partial.FullWidth != nil {
		config.FullWidth = *partial.FullWidth
	}
	if partial.FullTUI != nil {
		config.FullTUI = *partial.FullTUI
	}
	if partial.Rainbow != nil {
		opt := parseRainbowOption(partial.Rainbow)
		config.Rainbow = opt
		config.RainbowBar = opt
	}
	if partial.RainbowBar != nil {
		opt := parseRainbowOption(partial.RainbowBar)
		config.RainbowBar = opt
		config.Rainbow = opt
	}
}

func parseRainbowOption(val any) domain.RainbowOption {
	switch v := val.(type) {
	case bool:
		return domain.RainbowOption{Enabled: v}
	case []any:
		var colors []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				colors = append(colors, s)
			}
		}
		return domain.RainbowOption{Enabled: true, Colors: colors}
	case []string:
		return domain.RainbowOption{Enabled: true, Colors: v}
	}
	return domain.RainbowOption{Enabled: true}
}

func parseBarFg(val any) []string {
	var colors []string
	switch v := val.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v != "" {
			colors = append(colors, v)
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					colors = append(colors, s)
				}
			}
		}
	case []string:
		for _, s := range v {
			s = strings.TrimSpace(s)
			if s != "" {
				colors = append(colors, s)
			}
		}
	}
	if len(colors) > 10 {
		colors = colors[:10]
	}
	if len(colors) == 0 {
		return domain.DefaultConfig().BarFg
	}
	return colors
}

func expandPath(value string) string {
	expanded := os.ExpandEnv(value)
	if expanded == "" {
		return expanded
	}
	if expanded == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return expanded
	}
	if strings.HasPrefix(expanded, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(expanded, "~/"))
		}
	}
	return expanded
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

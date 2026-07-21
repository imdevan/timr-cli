package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/timr/internal/domain"
)

// Theme holds configurable colors for UI output.
type Theme struct {
	// New config colors
	TimeRemaining lipgloss.Color
	TimeStart     lipgloss.Color
	BarBg         lipgloss.Color
	BarFg         lipgloss.Color
	BarFgColors   []lipgloss.Color
	HelpText      lipgloss.Color
	Border        lipgloss.Color

	// Old colors kept for backward compatibility with other UI components
	Headings             lipgloss.Color
	Primary              lipgloss.Color
	Secondary            lipgloss.Color
	Text                 lipgloss.Color
	TextHighlight        lipgloss.Color
	DescriptionHighlight lipgloss.Color
	Tags                 lipgloss.Color
	Flags                lipgloss.Color
	Muted                lipgloss.Color

	// Feature flags
	RainbowBar    bool
	RainbowColors []lipgloss.Color
	ShowBorder    bool
	ShowHelpText  bool
}

// ThemeFromConfig builds a theme with safe fallbacks.
func ThemeFromConfig(cfg domain.Config) Theme {
	timeRemaining := resolveColor(cfg.TimeRemaining, "14")
	timeStart := resolveColor(cfg.TimeStart, "07")
	barBg := resolveColor(cfg.BarBg, "08")
	helpText := resolveColor(cfg.HelpText, "08")
	border := resolveColor(cfg.Border, "08")

	var barFgColors []lipgloss.Color
	for _, c := range cfg.BarFg {
		trimmed := strings.TrimSpace(c)
		if trimmed != "" {
			barFgColors = append(barFgColors, resolveColor(trimmed, domain.DefaultConfig().BarFg[0]))
		}
	}
	if len(barFgColors) == 0 {
		for _, c := range domain.DefaultConfig().BarFg {
			barFgColors = append(barFgColors, resolveColor(c, domain.DefaultConfig().BarFg[0]))
		}
	}
	barFg := barFgColors[0]

	rawColors := cfg.Rainbow.Colors
	if len(rawColors) == 0 {
		rawColors = cfg.RainbowBar.Colors
	}
	var customRainbow []lipgloss.Color
	for _, c := range rawColors {
		trimmed := strings.TrimSpace(c)
		if trimmed != "" {
			customRainbow = append(customRainbow, lipgloss.Color(trimmed))
		}
	}

	return Theme{
		TimeRemaining: timeRemaining,
		TimeStart:     timeStart,
		BarBg:         barBg,
		BarFg:         barFg,
		BarFgColors:   barFgColors,
		HelpText:      helpText,
		Border:        border,
		RainbowBar:    cfg.Rainbow.Enabled && cfg.RainbowBar.Enabled,
		RainbowColors: customRainbow,
		ShowBorder:    cfg.ShowBorder,
		ShowHelpText:  cfg.ShowHelpText,

		// Maps for backward compatibility
		Headings:             timeRemaining,
		Primary:              barFg,
		Secondary:            barFg,
		Text:                 timeStart,
		TextHighlight:        timeRemaining,
		DescriptionHighlight: timeRemaining,
		Tags:                 barFg,
		Flags:                timeRemaining,
		Muted:                helpText,
	}
}

func resolveColor(value, fallback string) lipgloss.Color {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = fallback
	}
	return lipgloss.Color(trimmed)
}

func resolveFallback(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

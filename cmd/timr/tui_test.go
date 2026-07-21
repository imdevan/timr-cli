package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/timr/internal/domain"
	"github.com/timr/internal/ui"
)

func maxLineWidth(s string) int {
	maxW := 0
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		w := lipgloss.Width(line)
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

func TestTimerModelFullWidth(t *testing.T) {
	cfg := domain.DefaultConfig()
	theme := ui.ThemeFromConfig(cfg)

	m := timerModel{
		duration:  10 * time.Minute,
		remaining: 5 * time.Minute,
		fullWidth: true,
		theme:     theme,
	}

	// Send WindowSizeMsg with width 80
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(timerModel)

	view := m.View()
	w := maxLineWidth(view)
	if w != 80 {
		t.Errorf("expected rendered width 80, got %d", w)
	}

	// Send WindowSizeMsg with width 100
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	m = updated.(timerModel)

	view = m.View()
	w = maxLineWidth(view)
	if w != 100 {
		t.Errorf("expected rendered width 100, got %d", w)
	}
}

func TestTimerModelFixedWidth(t *testing.T) {
	cfg := domain.DefaultConfig()
	theme := ui.ThemeFromConfig(cfg)

	m := timerModel{
		duration:  10 * time.Minute,
		remaining: 5 * time.Minute,
		fullWidth: false,
		theme:     theme,
	}

	// Send WindowSizeMsg with width 80
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(timerModel)

	view := m.View()
	w := maxLineWidth(view)
	// 40 inner width + 6 border & padding = 46
	if w != 46 {
		t.Errorf("expected rendered width 46 when fullWidth=false, got %d", w)
	}
}

func TestDoneModelFullWidth(t *testing.T) {
	cfg := domain.DefaultConfig()
	theme := ui.ThemeFromConfig(cfg)

	d := doneModel{
		fullWidth: true,
		theme:     theme,
	}

	// Send WindowSizeMsg with width 80
	updated, _ := d.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	d = updated.(doneModel)

	view := d.View()
	w := maxLineWidth(view)
	if w != 80 {
		t.Errorf("expected doneModel rendered width 80, got %d", w)
	}
}

func TestDoneModelFixedWidth(t *testing.T) {
	cfg := domain.DefaultConfig()
	theme := ui.ThemeFromConfig(cfg)

	d := doneModel{
		fullWidth: false,
		theme:     theme,
	}

	// Send WindowSizeMsg with width 80
	updated, _ := d.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	d = updated.(doneModel)

	view := d.View()
	w := maxLineWidth(view)
	// 40 inner width + 4 border & padding = 44
	if w != 44 {
		t.Errorf("expected doneModel rendered width 44 when fullWidth=false, got %d", w)
	}
}

func TestBarFgColorForTime(t *testing.T) {
	colors := []lipgloss.Color{
		lipgloss.Color("02"), // green (1st division)
		lipgloss.Color("03"), // yellow (2nd division)
		lipgloss.Color("01"), // red (3rd division)
	}
	duration := 60 * time.Second

	tests := []struct {
		remaining time.Duration
		wantColor lipgloss.Color
	}{
		{60 * time.Second, colors[0]}, // 100% remaining -> division 1 (green)
		{41 * time.Second, colors[0]}, // ~68% remaining -> division 1 (green)
		{40 * time.Second, colors[1]}, // 66.6% remaining -> division 2 (yellow)
		{21 * time.Second, colors[1]}, // ~35% remaining -> division 2 (yellow)
		{20 * time.Second, colors[2]}, // 33.3% remaining -> division 3 (red)
		{0 * time.Second, colors[2]},  // 0% remaining -> division 3 (red)
	}

	for _, tt := range tests {
		got := BarFgColorForTime(tt.remaining, duration, colors)
		if got != tt.wantColor {
			t.Errorf("BarFgColorForTime(%v, %v) = %v, want %v", tt.remaining, duration, got, tt.wantColor)
		}
	}
}

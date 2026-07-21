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

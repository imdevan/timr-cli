package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/timr/internal/ui"
)

type timerModel struct {
	duration      time.Duration
	remaining     time.Duration
	lastTickTime  time.Time
	endTime       time.Time
	paused        bool
	isMonitor     bool
	quitting      bool
	cancelled     bool
	theme         ui.Theme
	alarmSound    string
	tickInterval  time.Duration
}

func (m timerModel) Init() tea.Cmd {
	return tick(m.tickInterval)
}

type tickMsg time.Time

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m timerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			m.cancelled = true
			return m, tea.Quit
		case " ":
			if !m.isMonitor {
				if m.paused {
					m.paused = false
					m.lastTickTime = time.Now()
				} else {
					m.paused = true
				}
			}
		}

	case tickMsg:
		if m.quitting {
			return m, nil
		}

		now := time.Now()
		if m.isMonitor {
			m.remaining = time.Until(m.endTime)
			if m.remaining <= 0 {
				m.remaining = 0
				m.quitting = true
				return m, tea.Quit
			}
		} else {
			if !m.paused {
				elapsed := now.Sub(m.lastTickTime)
				m.remaining -= elapsed
				if m.remaining <= 0 {
					m.remaining = 0
					m.quitting = true
					go playAlarm(m.alarmSound)
					return m, tea.Quit
				}
			}
			m.lastTickTime = now
		}

		return m, tick(m.tickInterval)
	}

	return m, nil
}

func (m timerModel) View() string {
	if m.quitting {
		if m.cancelled {
			return lipgloss.NewStyle().Foreground(m.theme.Muted).Render("✗ Timer cancelled.\n")
		}
		return lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Render("⏰ Time's up!\n")
	}

	var sb strings.Builder

	// Title
	var title string
	if m.isMonitor {
		title = fmt.Sprintf("Monitoring Background Timer (%s)", formatDuration(m.duration))
	} else {
		title = fmt.Sprintf("Timer: %s", formatDuration(m.duration))
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(m.theme.Headings).Bold(true).Render(title))
	sb.WriteString("\n")

	// Countdown
	remainingStr := formatDuration(m.remaining)
	countdownStyle := lipgloss.NewStyle().
		Foreground(m.theme.TextHighlight).
		Bold(true)
	
	if m.paused {
		sb.WriteString(countdownStyle.Render(remainingStr + " [PAUSED]"))
	} else {
		sb.WriteString(countdownStyle.Render(remainingStr))
	}
	sb.WriteString("\n")

	// Progress Bar
	width := 40
	pct := float64(m.remaining) / float64(m.duration)
	if pct > 1.0 {
		pct = 1.0
	} else if pct < 0.0 {
		pct = 0.0
	}
	filledLen := int(pct * float64(width))
	if filledLen < 0 {
		filledLen = 0
	} else if filledLen > width {
		filledLen = width
	}
	emptyLen := width - filledLen

	filledBar := strings.Repeat("█", filledLen)
	emptyBar := strings.Repeat("░", emptyLen)

	barStr := lipgloss.NewStyle().Foreground(m.theme.Primary).Render(filledBar) +
		lipgloss.NewStyle().Foreground(m.theme.Muted).Render(emptyBar)
	
	sb.WriteString(barStr)
	sb.WriteString("\n")

	// Help / controls
	var helpStr string
	if m.isMonitor {
		helpStr = "[q/Esc] exit monitoring"
	} else {
		helpStr = "[Space] pause/resume • [q/Esc] cancel"
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(m.theme.Muted).Render(helpStr))
	sb.WriteString("\n")

	return sb.String()
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func playAlarm(soundPath string) {
	if soundPath == "" {
		beepTerminal()
		return
	}

	players := [][]string{
		{"mpv", "--no-terminal", soundPath},
		{"paplay", soundPath},
		{"aplay", soundPath},
		{"play", soundPath},
		{"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", soundPath},
	}

	played := false
	for _, args := range players {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Start(); err == nil {
			go func(c *exec.Cmd) {
				_ = c.Wait()
			}(cmd)
			played = true
			break
		}
	}

	if !played {
		beepTerminal()
	}
}

func beepTerminal() {
	for i := 0; i < 5; i++ {
		_, _ = os.Stdout.Write([]byte("\a"))
		time.Sleep(300 * time.Millisecond)
	}
}

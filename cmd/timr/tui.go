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
	confirmMode   bool
	confirmModel  *ui.ConfirmationModel
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
	if m.confirmMode && m.confirmModel != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			updated, cmd := m.confirmModel.Update(msg)
			if updatedConfirm, ok := updated.(ui.ConfirmationModel); ok {
				m.confirmModel = &updatedConfirm
				if cmd != nil {
					if _, isQuit := cmd().(tea.QuitMsg); isQuit {
						confirmed := m.confirmModel.ChoiceValue()
						m.confirmMode = false
						if confirmed {
							m.quitting = true
							m.cancelled = true
							return m, tea.Quit
						} else {
							m.lastTickTime = time.Now()
							return m, tick(m.tickInterval)
						}
					}
				}
			}
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			if !m.isMonitor && m.remaining > 0 {
				confirmModel := ui.NewConfirmationModel(
					"Cancel Timer",
					"Are you sure you want to cancel the timer?",
					m.theme,
				)
				m.confirmModel = &confirmModel
				m.confirmMode = true
				return m, confirmModel.Init()
			}
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
	if m.confirmMode && m.confirmModel != nil {
		return m.confirmModel.View()
	}

	if m.quitting {
		if m.cancelled {
			return lipgloss.NewStyle().Foreground(m.theme.Muted).Render("✗ Timer cancelled.\n")
		}
		return lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Render("⏰ Time's up!\n")
	}

	// 1. Build the first line: remaining time on left, total duration on right
	remStr := formatDuration(m.remaining)
	if m.paused {
		remStr = remStr + " [PAUSED]"
	}
	totStr := formatDuration(m.duration)

	visibleLen := len(remStr) + len(totStr)
	spaceCount := 40 - visibleLen
	if spaceCount < 1 {
		spaceCount = 1
	}
	spaces := strings.Repeat(" ", spaceCount)

	styledRem := lipgloss.NewStyle().Foreground(m.theme.TextHighlight).Bold(true).Render(remStr)
	styledTot := lipgloss.NewStyle().Foreground(m.theme.Muted).Render(totStr)
	firstLine := styledRem + spaces + styledTot

	// 2. Build the progress bar (width 40)
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

	// 3. Build help / controls
	var helpStr string
	if m.isMonitor {
		helpStr = "[q/Esc] exit monitoring"
	} else {
		helpStr = "[Space] pause/resume • [q/Esc] cancel"
	}
	styledHelp := lipgloss.NewStyle().Foreground(m.theme.Muted).Render(helpStr)

	// 4. Combine inner view
	inner := firstLine + "\n" + barStr + "\n" + styledHelp

	// 5. Wrap with a border using lipgloss
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(0, 1)

	return borderStyle.Render(inner) + "\n"
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

var activePlayCmd *exec.Cmd

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
		activePlayCmd = cmd
		if err := cmd.Start(); err == nil {
			_ = cmd.Wait()
			played = true
			break
		}
	}

	if !played {
		beepTerminal()
	}
}

func startPlayAlarmCmd(soundPath string) *exec.Cmd {
	if soundPath == "" {
		go beepTerminal()
		return nil
	}

	players := [][]string{
		{"mpv", "--no-terminal", soundPath},
		{"paplay", soundPath},
		{"aplay", soundPath},
		{"play", soundPath},
		{"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", soundPath},
	}

	for _, args := range players {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Start(); err == nil {
			return cmd
		}
	}

	go beepTerminal()
	return nil
}

func beepTerminal() {
	for i := 0; i < 5; i++ {
		_, _ = os.Stdout.Write([]byte("\a"))
		time.Sleep(300 * time.Millisecond)
	}
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/timr/internal/adapters/tmux"
	timeremaining "github.com/timr/internal/time_remaining"
	"github.com/timr/internal/ui"
)

type timerModel struct {
	duration           time.Duration
	remaining          time.Duration
	lastTickTime       time.Time
	endTime            time.Time
	paused             bool
	isMonitor          bool
	quitting           bool
	cancelled          bool
	theme              ui.Theme
	alarmSound         string
	tickInterval       time.Duration
	confirmMode        bool
	confirmModel       *ui.ConfirmationModel
	updateTmux         bool
	tmuxProgressBar    bool
	originalTmuxWindow string
	lastTmuxSeconds    int
	rainbowBar         bool
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
							if m.updateTmux && m.originalTmuxWindow != "" {
								setTmuxWindowName(m.originalTmuxWindow)
							}
							return m, tea.Quit
						} else {
							m.lastTickTime = time.Now()
							if m.updateTmux && os.Getenv("TMUX") != "" {
								setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar))
							}
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
			if m.updateTmux && m.originalTmuxWindow != "" {
				setTmuxWindowName(m.originalTmuxWindow)
			}
			return m, tea.Quit
		case " ":
			if !m.isMonitor {
				if m.paused {
					m.paused = false
					m.lastTickTime = time.Now()
					if m.updateTmux && os.Getenv("TMUX") != "" {
						setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar))
					}
				} else {
					m.paused = true
					if m.updateTmux && os.Getenv("TMUX") != "" {
						setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar))
					}
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
					if m.updateTmux && m.originalTmuxWindow != "" {
						setTmuxWindowName("⏰ done!")
					}
					return m, tea.Quit
				}

				if m.updateTmux && os.Getenv("TMUX") != "" {
					remSec := int(m.remaining.Round(time.Second).Seconds())
					if remSec != m.lastTmuxSeconds {
						m.lastTmuxSeconds = remSec
						setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar))
					}
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
			return lipgloss.NewStyle().Foreground(m.theme.HelpText).Render("✗ Timer cancelled.\n")
		}
		return ""
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

	styledRem := lipgloss.NewStyle().Foreground(m.theme.TimeRemaining).Bold(true).Render(remStr)
	styledTot := lipgloss.NewStyle().Foreground(m.theme.TimeStart).Render(totStr)
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

	barStr := lipgloss.NewStyle().Foreground(m.theme.BarFg).Render(filledBar) +
		lipgloss.NewStyle().Foreground(m.theme.BarBg).Render(emptyBar)

	// 3. Build help / controls
	var helpStr string
	if m.isMonitor {
		helpStr = "[q/Esc] exit monitoring"
	} else {
		helpStr = "[Space] pause/resume • [q/Esc] cancel"
	}
	styledHelp := lipgloss.NewStyle().Foreground(m.theme.HelpText).Render(helpStr)

	// 4. Combine inner view
	inner := firstLine + "\n" + barStr + "\n" + styledHelp

	// 5. Wrap with a border using lipgloss
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(1, 2)

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

// rainbowColors is the sequence of hues cycled across the bar.
var rainbowColors = []lipgloss.Color{
	"#f5bde6", // Pink
	"#c6a0f6", // Mauve
	"#ed8796", // Red
	"#ee99a0", // Maroon
	"#f5a97f", // Peach
	"#eed49f", // Yellow
	"#a6da95", // Green
	"#8bd5ca", // Teal
	"#91d7e3", // Sky
	"#7dc4e4", // Sapphire
	"#8aadf4", // Blue
	"#b7bdf8", // Lavender
}

// doneModel is a short-lived Bubble Tea program shown after the timer
// completes. It animates an oscillating rainbow bar inside the themed border
// while the alarm is playing, and exits on any keypress.
type doneModel struct {
	theme  ui.Theme
	phase  int
	stopCh <-chan struct{}
}

type doneTickMsg struct{}

func (d doneModel) Init() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return doneTickMsg{}
	})
}

func (d doneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return d, tea.Quit
	case doneTickMsg:
		// Check if the alarm finished (stopCh closed).
		select {
		case <-d.stopCh:
			return d, tea.Quit
		default:
		}
		d.phase++
		return d, tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
			return doneTickMsg{}
		})
	}
	return d, nil
}

func (d doneModel) View() string {
	const width = 40
	timesUpLine := lipgloss.NewStyle().
		Foreground(d.theme.TimeRemaining).
		Bold(true).
		Render("⏰ Time's up!")

	var barStr string
	if d.theme.RainbowBar {
		colors := d.theme.RainbowColors
		if len(colors) == 0 {
			colors = rainbowColors
		}
		n := len(colors)
		bar := make([]string, width)
		for i := range bar {
			// Oscillate: shift the hue offset sinusoidally using phase.
			colorIdx := (i + d.phase) % n
			bar[i] = lipgloss.NewStyle().
				Foreground(colors[colorIdx]).
				Render("█")
		}
		barStr = strings.Join(bar, "")
	} else {
		barStr = strings.Repeat(" ", width)
	}

	helpLine := lipgloss.NewStyle().
		Foreground(d.theme.HelpText).
		Render("Playing alarm... [Press any key to stop]")

	inner := timesUpLine + "\n" + barStr + "\n" + helpLine

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(d.theme.Border).
		Padding(0, 1)

	return borderStyle.Render(inner) + "\n"
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

var tmuxAdapter = tmux.New()

func getTmuxWindowName() (string, error) {
	return tmuxAdapter.OriginalName(), nil
}

func setTmuxWindowName(name string) {
	_ = tmuxAdapter.RenameWindow(name)
}

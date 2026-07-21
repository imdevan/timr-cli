package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/timr/internal/adapters/tmux"
	"github.com/timr/internal/domain"
	timeremaining "github.com/timr/internal/time_remaining"
	"github.com/timr/internal/ui"
)

type confirmActionType int

const (
	confirmActionCancel confirmActionType = iota
	confirmActionReset
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
	confirmAction      confirmActionType
	confirmModel       *ui.ConfirmationModel
	updateTmux         bool
	tmuxProgressBar    bool
	tmuxInverted       bool
	originalTmuxWindow string
	lastTmuxSeconds    int
	rainbowBar         bool
	phase              int
	fullWidth          bool
	fullTUI            bool
	vertical           bool
	verticalWidth      int
	pomodoroProgress   string
	termWidth          int
	termHeight         int
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
	if wmsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.termWidth = wmsg.Width
		m.termHeight = wmsg.Height
	}

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
						if m.confirmAction == confirmActionReset {
							if confirmed {
								m.remaining = m.duration
								m.lastTickTime = time.Now()
								if m.updateTmux && os.Getenv("TMUX") != "" {
									setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar, m.tmuxInverted))
								}
							} else {
								m.lastTickTime = time.Now()
							}
							return m, tick(m.tickInterval)
						} else {
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
									setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar, m.tmuxInverted))
								}
								return m, tick(m.tickInterval)
							}
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
				m.confirmAction = confirmActionCancel
				return m, confirmModel.Init()
			}
			m.quitting = true
			m.cancelled = true
			if m.updateTmux && m.originalTmuxWindow != "" {
				setTmuxWindowName(m.originalTmuxWindow)
			}
			return m, tea.Quit
		case "r":
			if !m.isMonitor {
				confirmModel := ui.NewConfirmationModel(
					"Reset Timer",
					"Are you sure you want to reset the timer?",
					m.theme,
				)
				m.confirmModel = &confirmModel
				m.confirmMode = true
				m.confirmAction = confirmActionReset
				return m, confirmModel.Init()
			}
		case " ":
			if !m.isMonitor {
				if m.paused {
					m.paused = false
					m.lastTickTime = time.Now()
					if m.updateTmux && os.Getenv("TMUX") != "" {
						setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar, m.tmuxInverted))
					}
				} else {
					m.paused = true
					if m.updateTmux && os.Getenv("TMUX") != "" {
						setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar, m.tmuxInverted))
					}
				}
			}
		}

	case tickMsg:
		if m.quitting {
			return m, nil
		}
		m.phase++

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
						setTmuxWindowName(timeremaining.Format(m.remaining, m.duration, m.paused, m.tmuxProgressBar, m.tmuxInverted))
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
		confirmView := m.confirmModel.View()
		if m.fullTUI && m.termWidth > 0 && m.termHeight > 0 {
			return ui.PlaceCenter(m.termWidth, m.termHeight, confirmView)
		}
		return confirmView
	}

	if m.quitting {
		if m.cancelled {
			return lipgloss.NewStyle().Foreground(m.theme.HelpText).Render("✗ Timer cancelled.\n")
		}
		return ""
	}

	width := 40
	var borderStyle lipgloss.Style
	if m.theme.ShowBorder {
		borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.theme.Border).
			Padding(1, 2)
	} else {
		borderStyle = lipgloss.NewStyle().
			Padding(1, 2)
	}

	if m.fullWidth && m.termWidth > 0 {
		innerWidth := m.termWidth - 6
		if innerWidth < 1 {
			innerWidth = 1
		}
		width = innerWidth
		boxWidth := m.termWidth - 2
		if boxWidth < 1 {
			boxWidth = 1
		}
		borderStyle = borderStyle.Width(boxWidth)
	}

	if m.vertical {
		remStr := formatDuration(m.remaining)
		if m.paused {
			remStr = remStr + " [PAUSED]"
		}
		totStr := formatDuration(m.duration)

		pct := float64(m.remaining) / float64(m.duration)
		if pct > 1.0 {
			pct = 1.0
		} else if pct < 0.0 {
			pct = 0.0
		}

		barHeight := 8
		if m.fullTUI && m.termHeight > 0 {
			// Overhead: padding (2) + trailing newline (1) = 3
			overhead := 3
			if m.theme.ShowBorder {
				overhead += 2 // border top + bottom
			}
			if m.theme.ShowHelpText {
				overhead += 3 // help gap (2) + help line (1)
			}
			availHeight := m.termHeight - overhead
			if availHeight > 3 {
				barHeight = availHeight
			}
		}

		vWidth := m.verticalWidth
		if vWidth < 1 {
			vWidth = 3
		}

		filledCount := int(pct * float64(barHeight))
		if filledCount < 0 {
			filledCount = 0
		} else if filledCount > barHeight {
			filledCount = barHeight
		}
		emptyCount := barHeight - filledCount

		bgStyle := lipgloss.NewStyle().Foreground(m.theme.BarBg)

		filledBlock := strings.Repeat("█", vWidth)
		emptyBlock := strings.Repeat("░", vWidth)

		var barLines []string
		// Empty rows on top, filled rows on bottom (bar fills upward)
		for i := 0; i < emptyCount; i++ {
			barLines = append(barLines, bgStyle.Render(emptyBlock))
		}
		barFgColor := BarFgColorForTime(m.remaining, m.duration, m.theme.BarFgColors)
		fgStyle := lipgloss.NewStyle().Foreground(barFgColor)
		for i := 0; i < filledCount; i++ {
			barLines = append(barLines, fgStyle.Render(filledBlock))
		}
		barVertical := strings.Join(barLines, "\n")

		styledTot := lipgloss.NewStyle().Foreground(m.theme.TimeStart).Render(totStr)
		styledRem := lipgloss.NewStyle().Foreground(m.theme.TimeRemaining).Bold(true).Render(remStr)
		var styledPom string
		leftWidth := lipgloss.Width(styledTot)
		if remW := lipgloss.Width(styledRem); remW > leftWidth {
			leftWidth = remW
		}
		if m.pomodoroProgress != "" {
			styledPom = lipgloss.NewStyle().Foreground(m.theme.TimeRemaining).Bold(true).Render(m.pomodoroProgress)
			if pomW := lipgloss.Width(styledPom); pomW > leftWidth {
				leftWidth = pomW
			}
		}

		centerStyle := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Center)
		styledTot = centerStyle.Render(styledTot)
		styledRem = centerStyle.Render(styledRem)
		if m.pomodoroProgress != "" {
			styledPom = centerStyle.Render(styledPom)
		}

		var leftContent string
		if m.pomodoroProgress != "" {
			bottomText := styledPom + "\n" + styledRem
			topPadCount := barHeight - 3
			if topPadCount < 1 {
				topPadCount = 1
			}
			leftContent = styledTot + strings.Repeat("\n", topPadCount) + bottomText
		} else {
			topPadCount := barHeight - 2
			if topPadCount < 1 {
				topPadCount = 1
			}
			leftContent = styledTot + strings.Repeat("\n", topPadCount) + styledRem
		}

		barCentered := lipgloss.NewStyle().Align(lipgloss.Center).Render(barVertical)
		mainBlock := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, "    ", barCentered)

		// Set border width so content can be centered horizontally
		if m.termWidth > 0 {
			borderInset := 2
			if !m.theme.ShowBorder {
				borderInset = 0
			}
			boxWidth := m.termWidth - borderInset
			if boxWidth < 1 {
				boxWidth = 1
			}
			borderStyle = borderStyle.Width(boxWidth)
		}

		innerWidth := lipgloss.Width(borderStyle.Render(""))
		if innerWidth < 1 {
			innerWidth = lipgloss.Width(mainBlock)
		}
		mainBlockCentered := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(mainBlock)

		var inner string
		if m.theme.ShowHelpText {
			var helpStr string
			if m.isMonitor {
				helpStr = "[q/Esc] exit monitoring"
			} else {
				helpStr = "[Space] pause/resume • [r] reset • [q/Esc] cancel"
			}
			styledHelp := lipgloss.NewStyle().Foreground(m.theme.HelpText).Render(helpStr)
			styledHelpCentered := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(styledHelp)
			inner = mainBlockCentered + "\n\n" + styledHelpCentered
		} else {
			inner = mainBlockCentered
		}

		view := borderStyle.Render(inner) + "\n"
		if m.fullTUI && m.termWidth > 0 && m.termHeight > 0 {
			return ui.PlaceCenter(m.termWidth, m.termHeight, view)
		}
		return view
	}

	// 1. Build the first line: remaining time on left, total duration on right
	remStr := formatDuration(m.remaining)
	if m.pomodoroProgress != "" {
		remStr = m.pomodoroProgress + " " + remStr
	}
	if m.paused {
		remStr = remStr + " [PAUSED]"
	}
	totStr := formatDuration(m.duration)

	visibleLen := len(remStr) + len(totStr)
	spaceCount := width - visibleLen
	if spaceCount < 1 {
		spaceCount = 1
	}
	spaces := strings.Repeat(" ", spaceCount)

	styledRem := lipgloss.NewStyle().Foreground(m.theme.TimeRemaining).Bold(true).Render(remStr)
	styledTot := lipgloss.NewStyle().Foreground(m.theme.TimeStart).Render(totStr)
	firstLine := styledRem + spaces + styledTot

	// 2. Build the progress bar
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

	barFgColor := BarFgColorForTime(m.remaining, m.duration, m.theme.BarFgColors)
	barStr := lipgloss.NewStyle().Foreground(barFgColor).Render(filledBar) +
		lipgloss.NewStyle().Foreground(m.theme.BarBg).Render(emptyBar)

	// 3. Build help / controls
	var inner string
	if m.theme.ShowHelpText {
		var helpStr string
		if m.isMonitor {
			helpStr = "[q/Esc] exit monitoring"
		} else {
			helpStr = "[Space] pause/resume • [r] reset • [q/Esc] cancel"
		}
		styledHelp := lipgloss.NewStyle().Foreground(m.theme.HelpText).Render(helpStr)
		inner = firstLine + "\n" + barStr + "\n" + styledHelp
	} else {
		inner = firstLine + "\n" + barStr
	}

	// 4. Wrap with a border using lipgloss
	view := borderStyle.Render(inner) + "\n"
	if m.fullTUI && m.termWidth > 0 && m.termHeight > 0 {
		return ui.PlaceVertically(m.termWidth, m.termHeight, view)
	}
	return view
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

// defaultRainbowAnchors is the keyframe sequence for the default rainbow.
var defaultRainbowAnchors = []lipgloss.Color{
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
	"#c6a0f6", // Mauve
	"#f5bde6", // Pink
	"#f0c6c6", // Flamingo
	"#f4dbd6", // Rosewater
}

func parseHexColor(hex string) (r, g, b float64, ok bool) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, 0, 0, false
	}
	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return float64((val >> 16) & 0xFF), float64((val >> 8) & 0xFF), float64(val & 0xFF), true
}

// generateCircularGradient builds a smooth closed-loop color palette
// with even interpolation between keyframes so the end fades into the beginning.
func generateCircularGradient(anchors []lipgloss.Color) []lipgloss.Color {
	n := len(anchors)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return []lipgloss.Color{anchors[0]}
	}

	stepsPerSegment := 4
	if n <= 4 {
		stepsPerSegment = 16 / n
		if stepsPerSegment < 4 {
			stepsPerSegment = 4
		}
	}

	var palette []lipgloss.Color
	for i := 0; i < n; i++ {
		c1Str := string(anchors[i])
		c2Str := string(anchors[(i+1)%n])

		r1, g1, b1, ok1 := parseHexColor(c1Str)
		r2, g2, b2, ok2 := parseHexColor(c2Str)

		if !ok1 || !ok2 {
			palette = append(palette, anchors[i])
			continue
		}

		for s := 0; s < stepsPerSegment; s++ {
			t := float64(s) / float64(stepsPerSegment)
			r := uint8(math.Round(r1 + t*(r2-r1)))
			g := uint8(math.Round(g1 + t*(g2-g1)))
			b := uint8(math.Round(b1 + t*(b2-b1)))
			hexStr := fmt.Sprintf("#%02x%02x%02x", r, g, b)
			palette = append(palette, lipgloss.Color(hexStr))
		}
	}
	return palette
}

// doneModel is a short-lived Bubble Tea program shown after the timer
// completes. It animates an oscillating rainbow bar inside the themed border
// while the alarm is playing, and exits on any keypress.
type doneModel struct {
	theme         ui.Theme
	phase         int
	stopCh        <-chan struct{}
	fullWidth     bool
	fullTUI       bool
	vertical      bool
	verticalWidth int
	termWidth     int
	termHeight    int
}

type doneTickMsg struct{}

func (d doneModel) Init() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return doneTickMsg{}
	})
}

func (d doneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wmsg, ok := msg.(tea.WindowSizeMsg); ok {
		d.termWidth = wmsg.Width
		d.termHeight = wmsg.Height
	}

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
	width := 40
	var borderStyle lipgloss.Style
	if d.theme.ShowBorder {
		borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(d.theme.Border).
			Padding(0, 1)
	} else {
		borderStyle = lipgloss.NewStyle().
			Padding(0, 1)
	}

	if d.fullWidth && d.termWidth > 0 {
		innerWidth := d.termWidth - 4
		if innerWidth < 1 {
			innerWidth = 1
		}
		width = innerWidth
		boxWidth := d.termWidth - 2
		if boxWidth < 1 {
			boxWidth = 1
		}
		borderStyle = borderStyle.Width(boxWidth)
	}

	if d.vertical {
		timesUpLine := lipgloss.NewStyle().
			Foreground(d.theme.TimeRemaining).
			Bold(true).
			Render("⏰ Time's up!")

		barHeight := 8
		if d.fullTUI && d.termHeight > 0 {
			overhead := 1 // trailing newline
			if d.theme.ShowBorder {
				overhead += 2 // border top + bottom
			}
			if d.theme.ShowHelpText {
				overhead += 3 // help gap (2) + help line (1)
			}
			availHeight := d.termHeight - overhead
			if availHeight > 3 {
				barHeight = availHeight
			}
		}

		vWidth := d.verticalWidth
		if vWidth < 1 {
			vWidth = 3
		}

		filledBlock := strings.Repeat("█", vWidth)
		emptyBlock := strings.Repeat(" ", vWidth)

		var barLines []string
		if d.theme.RainbowBar {
			anchors := d.theme.RainbowColors
			if len(anchors) == 0 {
				anchors = defaultRainbowAnchors
			}
			colors := generateCircularGradient(anchors)
			n := len(colors)
			// Rainbow fills vertically from bottom to top
			for i := 0; i < barHeight; i++ {
				// Map from bottom: row 0 (top) = barHeight-1, last row = 0
				bottomIdx := barHeight - 1 - i
				colorIdx := ((bottomIdx / 2) + d.phase) % n
				barLines = append(barLines, lipgloss.NewStyle().Foreground(colors[colorIdx]).Render(filledBlock))
			}
		} else {
			bgStyle := lipgloss.NewStyle().Foreground(d.theme.BarBg)
			for i := 0; i < barHeight; i++ {
				barLines = append(barLines, bgStyle.Render(emptyBlock))
			}
		}
		barVertical := strings.Join(barLines, "\n")

		leftContent := timesUpLine
		topPadCount := barHeight - 2
		if topPadCount < 1 {
			topPadCount = 1
		}
		leftContent = timesUpLine + strings.Repeat("\n", topPadCount)

		mainBlock := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, "    ", barVertical)

		// Set border width so content can be centered horizontally
		if d.termWidth > 0 {
			borderInset := 2
			if !d.theme.ShowBorder {
				borderInset = 0
			}
			boxWidth := d.termWidth - borderInset
			if boxWidth < 1 {
				boxWidth = 1
			}
			borderStyle = borderStyle.Width(boxWidth)
		}

		innerWidth := lipgloss.Width(borderStyle.Render(""))
		if innerWidth < 1 {
			innerWidth = lipgloss.Width(mainBlock)
		}
		mainBlockCentered := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(mainBlock)

		var inner string
		if d.theme.ShowHelpText {
			helpLine := lipgloss.NewStyle().
				Foreground(d.theme.HelpText).
				Render("Playing alarm... [Press any key to stop]")
			helpCentered := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(helpLine)
			inner = mainBlockCentered + "\n\n" + helpCentered
		} else {
			inner = mainBlockCentered
		}

		view := borderStyle.Render(inner) + "\n"
		if d.fullTUI && d.termWidth > 0 && d.termHeight > 0 {
			return ui.PlaceCenter(d.termWidth, d.termHeight, view)
		}
		return view
	}

	timesUpLine := lipgloss.NewStyle().
		Foreground(d.theme.TimeRemaining).
		Bold(true).
		Render("⏰ Time's up!")

	var barStr string
	if d.theme.RainbowBar {
		anchors := d.theme.RainbowColors
		if len(anchors) == 0 {
			anchors = defaultRainbowAnchors
		}
		colors := generateCircularGradient(anchors)
		n := len(colors)
		bar := make([]string, width)
		for i := range bar {
			// Oscillate: shift hue with 2 bars (characters) per color step.
			colorIdx := ((i / 2) + d.phase) % n
			bar[i] = lipgloss.NewStyle().
				Foreground(colors[colorIdx]).
				Render("█")
		}
		barStr = strings.Join(bar, "")
	} else {
		barStr = strings.Repeat(" ", width)
	}

	var inner string
	if d.theme.ShowHelpText {
		helpLine := lipgloss.NewStyle().
			Foreground(d.theme.HelpText).
			Render("Playing alarm... [Press any key to stop]")
		inner = timesUpLine + "\n" + barStr + "\n" + helpLine
	} else {
		inner = timesUpLine + "\n" + barStr
	}

	view := borderStyle.Render(inner) + "\n"
	if d.fullTUI && d.termWidth > 0 && d.termHeight > 0 {
		return ui.PlaceVertically(d.termWidth, d.termHeight, view)
	}
	return view
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

func BarFgColorForTime(remaining, duration time.Duration, colors []lipgloss.Color) lipgloss.Color {
	if len(colors) == 0 {
		return lipgloss.Color(domain.DefaultConfig().BarFg[0])
	}
	if len(colors) == 1 || duration <= 0 {
		return colors[0]
	}
	pct := float64(remaining) / float64(duration)
	if pct > 1.0 {
		pct = 1.0
	} else if pct < 0.0 {
		pct = 0.0
	}
	n := float64(len(colors))
	elapsed := 1.0 - pct
	idx := int(elapsed * n)
	if idx >= len(colors) {
		idx = len(colors) - 1
	} else if idx < 0 {
		idx = 0
	}
	return colors[idx]
}

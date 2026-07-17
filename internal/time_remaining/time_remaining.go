package timeremaining

import (
	"fmt"
	"time"
)

// Format returns the formatted string for the tmux window.
func Format(remaining time.Duration, paused bool) string {
	h := int(remaining.Hours())
	m := int(remaining.Minutes()) % 60
	s := int(remaining.Seconds()) % 60

	var remStr string
	if h > 0 {
		remStr = fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	} else {
		remStr = fmt.Sprintf("%02d:%02d", m, s)
	}

	if paused {
		return fmt.Sprintf("⏰ %s [PAUSED]", remStr)
	}
	return fmt.Sprintf("⏰ %s", remStr)
}

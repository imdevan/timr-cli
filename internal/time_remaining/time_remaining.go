package timeremaining

import (
	"fmt"
	"time"
)

var moonPhases = []rune{
	'', // 100% remaining: weather-moon_alt_new (\ue3eb)
	'', // 93% remaining: weather-moon_alt_waxing_crescent_1 (\ue3d0)
	'', // 86% remaining: weather-moon_alt_waxing_crescent_2 (\ue3d1)
	'', // 79% remaining: weather-moon_alt_waxing_crescent_3 (\ue3d2)
	'', // 72% remaining: weather-moon_alt_waxing_crescent_4 (\ue3d3)
	'', // 65% remaining: weather-moon_alt_waxing_crescent_5 (\ue3d4)
	'', // 58% remaining: weather-moon_alt_waxing_crescent_6 (\ue3d5)
	'', // 51% remaining: weather-moon_alt_first_quarter (\ue3d6)
	'', // 44% remaining: weather-moon_alt_waxing_gibbous_1 (\ue3d7)
	'', // 37% remaining: weather-moon_alt_waxing_gibbous_2 (\ue3d8)
	'', // 30% remaining: weather-moon_alt_waxing_gibbous_3 (\ue3d9)
	'', // 23% remaining: weather-moon_alt_waxing_gibbous_4 (\ue3da)
	'', // 16% remaining: weather-moon_alt_waxing_gibbous_5 (\ue3db)
	'', // 9% remaining: weather-moon_alt_waxing_gibbous_6 (\ue3dc)
	'', // 2% remaining (and below): weather-moon_alt_full (\ue3dd)
}

// Format returns the formatted string for the tmux window.
func Format(remaining, total time.Duration, paused bool, showProgressBar bool) string {
	h := int(remaining.Hours())
	m := int(remaining.Minutes()) % 60
	s := int(remaining.Seconds()) % 60

	var remStr string
	if h > 0 {
		remStr = fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	} else {
		remStr = fmt.Sprintf("%02d:%02d", m, s)
	}

	var prefix string
	if showProgressBar && total > 0 {
		pct := float64(remaining) / float64(total)
		if pct > 1.0 {
			pct = 1.0
		} else if pct < 0.0 {
			pct = 0.0
		}
		pctRemaining := pct * 100.0
		var idx int
		if pctRemaining >= 96.5 {
			idx = 0
		} else if pctRemaining >= 89.5 {
			idx = 1
		} else if pctRemaining >= 82.5 {
			idx = 2
		} else if pctRemaining >= 75.5 {
			idx = 3
		} else if pctRemaining >= 68.5 {
			idx = 4
		} else if pctRemaining >= 61.5 {
			idx = 5
		} else if pctRemaining >= 54.5 {
			idx = 6
		} else if pctRemaining >= 47.5 {
			idx = 7
		} else if pctRemaining >= 40.5 {
			idx = 8
		} else if pctRemaining >= 33.5 {
			idx = 9
		} else if pctRemaining >= 26.5 {
			idx = 10
		} else if pctRemaining >= 19.5 {
			idx = 11
		} else if pctRemaining >= 12.5 {
			idx = 12
		} else if pctRemaining >= 5.5 {
			idx = 13
		} else {
			idx = 14
		}
		prefix = string(moonPhases[idx]) + " "
	} else {
		prefix = "⏰ "
	}

	if paused {
		return fmt.Sprintf("%s%s [PAUSED]", prefix, remStr)
	}
	return fmt.Sprintf("%s%s", prefix, remStr)
}

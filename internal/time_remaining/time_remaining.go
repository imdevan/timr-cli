package timeremaining

import (
	"fmt"
	"time"
)

var standardMoonPhases = []rune{
	'\ue3e3', // nf-weather-moon_alt_new (\ue3e3)
	'\ue3c8', // nf-weather-moon_alt_waxing_crescent_1 (\ue3c8)
	'\ue3c9', // nf-weather-moon_alt_waxing_crescent_2 (\ue3c9)
	'\ue3ca', // nf-weather-moon_alt_waxing_crescent_3 (\ue3ca)
	'\ue3cb', // nf-weather-moon_alt_waxing_crescent_4 (\ue3cb)
	'\ue3cc', // nf-weather-moon_alt_waxing_crescent_5 (\ue3cc)
	'\ue3cd', // nf-weather-moon_alt_waxing_crescent_6 (\ue3cd)
	'\ue3ce', // nf-weather-moon_alt_first_quarter (\ue3ce)
	'\ue3cf', // nf-weather-moon_alt_waxing_gibbous_1 (\ue3cf)
	'\ue3d0', // nf-weather-moon_alt_waxing_gibbous_2 (\ue3d0)
	'\ue3d1', // nf-weather-moon_alt_waxing_gibbous_3 (\ue3d1)
	'\ue3d2', // nf-weather-moon_alt_waxing_gibbous_4 (\ue3d2)
	'\ue3d3', // nf-weather-moon_alt_waxing_gibbous_5 (\ue3d3)
	'\ue3d4', // nf-weather-moon_alt_waxing_gibbous_6 (\ue3d4)
	'\ue3d5', // nf-weather-moon_alt_full (\ue3d5)
}

var invertedMoonPhases = []rune{
	'\ue3d5', // nf-weather-moon_alt_full (\ue3d5)
	'\ue3d6', // nf-weather-moon_alt_waning_gibbous_1 (\ue3d6)
	'\ue3d7', // nf-weather-moon_alt_waning_gibbous_2 (\ue3d7)
	'\ue3d8', // nf-weather-moon_alt_waning_gibbous_3 (\ue3d8)
	'\ue3d9', // nf-weather-moon_alt_waning_gibbous_4 (\ue3d9)
	'\ue3da', // nf-weather-moon_alt_waning_gibbous_5 (\ue3da)
	'\ue3db', // nf-weather-moon_alt_waning_gibbous_6 (\ue3db)
	'\ue3dc', // nf-weather-moon_alt_third_quarter (\ue3dc)
	'\ue3dd', // nf-weather-moon_alt_waning_crescent_1 (\ue3dd)
	'\ue3de', // nf-weather-moon_alt_waning_crescent_2 (\ue3de)
	'\ue3df', // nf-weather-moon_alt_waning_crescent_3 (\ue3df)
	'\ue3e0', // nf-weather-moon_alt_waning_crescent_4 (\ue3e0)
	'\ue3e1', // nf-weather-moon_alt_waning_crescent_5 (\ue3e1)
	'\ue3e2', // nf-weather-moon_alt_waning_crescent_6 (\ue3e2)
	'\ue3e3', // nf-weather-moon_alt_new (\ue3e3)
}

// Format returns the formatted string for the tmux window.
func Format(remaining, total time.Duration, paused bool, showProgressBar bool, inverted bool) string {
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
		phases := standardMoonPhases
		if inverted {
			phases = invertedMoonPhases
		}
		prefix = string(phases[idx]) + " "
	} else {
		prefix = "⏰ "
	}

	if paused {
		return fmt.Sprintf("%s%s [PAUSED]", prefix, remStr)
	}
	return fmt.Sprintf("%s%s", prefix, remStr)
}

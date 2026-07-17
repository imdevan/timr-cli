package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses duration strings with support for raw numbers (using defaultUnits),
// standard Go durations (10h, 10m, 10s, 10.5h), and colons (h:m, h:m:s).
func ParseDuration(input string, defaultUnits string) (time.Duration, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// Case 1: Colons format (e.g. h:m:s or h:m)
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		if len(parts) == 3 {
			// h:m:s
			hours, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid hours: %s", parts[0])
			}
			minutes, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid minutes: %s", parts[1])
			}
			seconds, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid seconds: %s", parts[2])
			}
			return time.Duration(hours*float64(time.Hour) + minutes*float64(time.Minute) + seconds*float64(time.Second)), nil
		} else if len(parts) == 2 {
			// h:m
			hours, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid hours: %s", parts[0])
			}
			minutes, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid minutes: %s", parts[1])
			}
			return time.Duration(hours*float64(time.Hour) + minutes*float64(time.Minute)), nil
		}
		return 0, fmt.Errorf("invalid colon format, use h:m or h:m:s")
	}

	// Case 2: Number with units (e.g. 10h, 10.5h, 10m, 10s)
	// We check if it ends with one of the duration suffixes
	lastChar := input[len(input)-1:]
	if lastChar == "s" || lastChar == "m" || lastChar == "h" {
		d, err := time.ParseDuration(input)
		if err == nil {
			return d, nil
		}
		return 0, err
	}

	// Case 3: Raw number (no units specified)
	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %s", input)
	}

	// Apply default units
	unit := strings.ToLower(strings.TrimSpace(defaultUnits))
	switch unit {
	case "s", "sec", "secs", "second", "seconds":
		return time.Duration(val * float64(time.Second)), nil
	case "h", "hour", "hours":
		return time.Duration(val * float64(time.Hour)), nil
	case "m", "min", "mins", "minute", "minutes", "":
		return time.Duration(val * float64(time.Minute)), nil
	default:
		return time.Duration(val * float64(time.Minute)), nil
	}
}

// TimeAgo returns a human-readable string representing how long ago the given time was.
func TimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

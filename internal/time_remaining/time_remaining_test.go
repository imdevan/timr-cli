package timeremaining

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name            string
		remaining       time.Duration
		total           time.Duration
		paused          bool
		showProgressBar bool
		want            string
	}{
		{
			name:            "running minute duration, no progress bar",
			remaining:       10 * time.Minute,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: false,
			want:            "⏰ 10:00",
		},
		{
			name:            "paused minute duration, no progress bar",
			remaining:       10*time.Minute + 15*time.Second,
			total:           10*time.Minute + 15*time.Second,
			paused:          true,
			showProgressBar: false,
			want:            "⏰ 10:15 [PAUSED]",
		},
		{
			name:            "running minute duration, with progress bar at 100%",
			remaining:       10 * time.Minute,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: true,
			want:            " 10:00", // weather-moon_alt_new (\ue3eb)
		},
		{
			name:            "running minute duration, with progress bar at 0%",
			remaining:       0,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: true,
			want:            " 00:00", // weather-moon_alt_full (\ue3dd)
		},
		{
			name:            "paused minute duration, with progress bar at 50%",
			remaining:       5 * time.Minute,
			total:           10 * time.Minute,
			paused:          true,
			showProgressBar: true,
			want:            " 05:00 [PAUSED]", // weather-moon_alt_first_quarter (\ue3d6, index 7)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.remaining, tt.total, tt.paused, tt.showProgressBar)
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

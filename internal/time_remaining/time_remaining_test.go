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
		inverted        bool
		want            string
	}{
		{
			name:            "running minute duration, no progress bar",
			remaining:       10 * time.Minute,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: false,
			inverted:        false,
			want:            "⏰ 10:00",
		},
		{
			name:            "paused minute duration, no progress bar",
			remaining:       10*time.Minute + 15*time.Second,
			total:           10*time.Minute + 15*time.Second,
			paused:          true,
			showProgressBar: false,
			inverted:        false,
			want:            "⏰ 10:15 [PAUSED]",
		},
		{
			name:            "running minute duration, standard progress bar at 100%",
			remaining:       10 * time.Minute,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: true,
			inverted:        false,
			want:            "\ue3e3 10:00", // nf-weather-moon_alt_new (\ue3e3)
		},
		{
			name:            "running minute duration, standard progress bar at 0%",
			remaining:       0,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: true,
			inverted:        false,
			want:            "\ue3d5 00:00", // nf-weather-moon_alt_full (\ue3d5)
		},
		{
			name:            "paused minute duration, standard progress bar at 50%",
			remaining:       5 * time.Minute,
			total:           10 * time.Minute,
			paused:          true,
			showProgressBar: true,
			inverted:        false,
			want:            "\ue3ce 05:00 [PAUSED]", // nf-weather-moon_alt_first_quarter (\ue3ce)
		},
		{
			name:            "running minute duration, inverted progress bar at 100%",
			remaining:       10 * time.Minute,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: true,
			inverted:        true,
			want:            "\ue3d5 10:00", // nf-weather-moon_alt_full (\ue3d5)
		},
		{
			name:            "running minute duration, inverted progress bar at 0%",
			remaining:       0,
			total:           10 * time.Minute,
			paused:          false,
			showProgressBar: true,
			inverted:        true,
			want:            "\ue3e3 00:00", // nf-weather-moon_alt_new (\ue3e3)
		},
		{
			name:            "paused minute duration, inverted progress bar at 50%",
			remaining:       5 * time.Minute,
			total:           10 * time.Minute,
			paused:          true,
			showProgressBar: true,
			inverted:        true,
			want:            "\ue3dc 05:00 [PAUSED]", // nf-weather-moon_alt_third_quarter (\ue3dc)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.remaining, tt.total, tt.paused, tt.showProgressBar, tt.inverted)
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

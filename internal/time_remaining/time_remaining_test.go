package timeremaining

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name      string
		remaining time.Duration
		paused    bool
		want      string
	}{
		{
			name:      "running minute duration",
			remaining: 10 * time.Minute,
			paused:    false,
			want:      "⏰ 10:00",
		},
		{
			name:      "paused minute duration",
			remaining: 10*time.Minute + 15*time.Second,
			paused:    true,
			want:      "⏰ 10:15 [PAUSED]",
		},
		{
			name:      "running hour duration",
			remaining: 1*time.Hour + 5*time.Minute + 30*time.Second,
			paused:    false,
			want:      "⏰ 01:05:30",
		},
		{
			name:      "paused hour duration",
			remaining: 2*time.Hour + 45*time.Second,
			paused:    true,
			want:      "⏰ 02:00:45 [PAUSED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.remaining, tt.paused)
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

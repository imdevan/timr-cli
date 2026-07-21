package ui

import (
	"testing"
)

func TestPlaceLayoutHelpers(t *testing.T) {
	content := "Hello World"
	t.Run("returns original content if dimensions are non-positive", func(t *testing.T) {
		if got := PlaceVertically(0, 0, content); got != content {
			t.Errorf("PlaceVertically(0, 0) = %q, want %q", got, content)
		}
		if got := PlaceCenter(0, 0, content); got != content {
			t.Errorf("PlaceCenter(0, 0) = %q, want %q", got, content)
		}
	})

	t.Run("places content in valid dimensions", func(t *testing.T) {
		vert := PlaceVertically(80, 24, content)
		if vert == content || len(vert) == 0 {
			t.Errorf("expected PlaceVertically to format content for window")
		}
		center := PlaceCenter(80, 24, content)
		if center == content || len(center) == 0 {
			t.Errorf("expected PlaceCenter to format content for window")
		}
	})
}

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// PlaceVertically centers content vertically in a window of size (width, height), aligned to the left.
func PlaceVertically(width, height int, content string) string {
	if width <= 0 || height <= 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Center, content)
}

// PlaceCenter centers content both horizontally and vertically in a window of size (width, height).
func PlaceCenter(width, height int, content string) string {
	if width <= 0 || height <= 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// fit trims a rendered line to the current terminal width.
func (m Model) fit(width int, line string) string {
	if width <= 0 {
		return line
	}

	if lipgloss.Width(line) <= width {
		return line
	}

	return lipgloss.NewStyle().MaxWidth(width).Render(line)
}

// padLines pads lines to exactly height rows.
func padLines(lines []string, height int) []string {
	for len(lines) < height {
		lines = append(lines, "")
	}

	if len(lines) > height {
		return lines[:height]
	}

	return lines
}

// isCoreField reports whether field is already rendered specially.
func isCoreField(field string) bool {
	switch strings.ToLower(field) {
	case "raw", "timestamp", "time", "ts", "level", "severity", "message", "msg":
		return true
	default:
		return false
	}
}

// padRight pads a string on the right to width.
func padRight(value string, width int) string {
	if len(value) >= width {
		return value
	}

	return value + strings.Repeat(" ", width-len(value))
}

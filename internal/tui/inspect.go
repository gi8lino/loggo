package tui

import (
	"fmt"
	"maps"
	"slices"
)

// inspectView renders the selected entry inspector.
func (m Model) inspectView(width int, height int) []string {
	entry, ok := m.activeEntry()
	if !ok {
		return padLines([]string{dimStyle.Render(" no entry selected")}, height)
	}

	lines := []string{
		"entry",
		"-----",
		fmt.Sprintf("parser      %s", entry.Parser),
		fmt.Sprintf("parsed      %t", entry.Parsed),
		fmt.Sprintf("timestamp   %s", entry.Timestamp),
		fmt.Sprintf("has_time    %t", entry.HasTime),
		fmt.Sprintf("level       %s", entry.Level),
		fmt.Sprintf("message     %s", entry.Message),
		"",
		"fields",
		"------",
	}

	keys := slices.Sorted(maps.Keys(entry.Fields))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%-12s %s", key, entry.Fields[key]))
	}

	lines = append(lines, "", "raw", "---", entry.Raw)

	for index, line := range lines {
		lines[index] = m.fit(width, line)
	}

	return padLines(lines, height)
}

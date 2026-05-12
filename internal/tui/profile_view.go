package tui

import (
	"fmt"
)

// profileView renders the profile picker.
func (m Model) profileView(width int, height int) []string {
	lines := []string{
		"switch profile",
		"--------------",
	}

	for index, name := range m.profileNames {
		prefix := "  "
		if index == m.profileCursor {
			prefix = "> "
		}

		p := m.profiles[name]
		line := fmt.Sprintf("%s%-20s parser=%s", prefix, name, p.Parser)

		if name == m.activeProfile.Name {
			line += " active"
		}

		lines = append(lines, line)
	}

	for index, line := range lines {
		lines[index] = m.fit(width, line)
	}

	return padLines(lines, height)
}

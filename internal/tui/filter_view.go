package tui

import (
	"fmt"
	"strings"
)

// filterFieldView renders the guided filter field picker.
func (m Model) filterFieldView(width int, height int) []string {
	fields := m.filterFieldOptions
	if len(fields) == 0 {
		fields = m.buildFilterFields()
	}

	title := "filter field"
	if m.mode == modeExcludeField {
		title = "exclude field"
	}

	lines := []string{
		title,
		strings.Repeat("-", len(title)),
	}

	if len(fields) == 0 {
		lines = append(lines, "no fields available")
		return padLines(lines, height)
	}

	start := max(0, m.filterFieldCursor-height+4)
	end := min(len(fields), start+height-2)

	for index := start; index < end; index++ {
		prefix := "  "
		if index == m.filterFieldCursor {
			prefix = "> "
		}

		lines = append(lines, prefix+fields[index])
	}

	for index, line := range lines {
		lines[index] = m.fit(width, line)
	}

	return padLines(lines, height)
}

// filterOperatorView renders the guided filter operator picker.
func (m Model) filterOperatorView(width int, height int) []string {
	title := fmt.Sprintf("field: %s", m.filterField)
	lines := []string{
		title,
		strings.Repeat("-", len(title)),
	}

	for index, operator := range guidedOperators {
		prefix := "  "
		if index == m.filterOperatorCursor {
			prefix = "> "
		}

		lines = append(lines, fmt.Sprintf("%s%s", prefix, operator.Label))
	}

	for index, line := range lines {
		lines[index] = m.fit(width, line)
	}

	return padLines(lines, height)
}

package tui

import "strings"

// columnsView renders the visible column picker.
func (m Model) columnsView(width int, height int) []string {
	fields := m.columnFieldOptions
	if len(fields) == 0 {
		fields = m.buildColumnFields()
	}

	hidden := m.columnHiddenDraft
	if hidden == nil {
		hidden = m.hiddenFields
	}

	title := "visible columns"
	lines := []string{
		title,
		strings.Repeat("-", len(title)),
	}

	if len(fields) == 0 {
		lines = append(lines, "no columns available")
		return padLines(lines, height)
	}

	start := max(0, m.columnFieldCursor-height+4)
	end := min(len(fields), start+height-2)

	for index := start; index < end; index++ {
		field := fields[index]
		prefix := "  "
		if index == m.columnFieldCursor {
			prefix = "> "
		}

		marker := "[x]"
		if _, ok := hidden[field]; ok {
			marker = "[ ]"
		}

		lines = append(lines, prefix+marker+" "+field)
	}

	for index, line := range lines {
		lines[index] = m.fit(width, line)
	}

	return padLines(lines, height)
}

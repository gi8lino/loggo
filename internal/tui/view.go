package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/gi8lino/loggo/internal/logentry"
)

const (
	timestampColumnWidth = 24
	levelColumnWidth     = 5
	fieldColumnWidth     = 18
)

// View renders the complete TUI.
func (m Model) View() tea.View {
	width := m.width
	if width <= 0 {
		width = 100
	}

	height := m.height
	if height <= 0 {
		height = 30
	}

	bodyHeight := height - 4
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	lines := []string{
		m.fit(width, m.statusLine(width)),
		m.fit(width, m.activeLine()),
	}

	switch m.mode {
	case modeHelp:
		lines = append(lines, m.helpView(width, bodyHeight)...)
	case modeInspect:
		lines = append(lines, m.inspectView(width, bodyHeight)...)
	case modeProfile:
		lines = append(lines, m.profileView(width, bodyHeight)...)
	case modeFilterField, modeExcludeField:
		lines = append(lines, m.filterFieldView(width, bodyHeight)...)
	case modeFilterOperator, modeExcludeOperator:
		lines = append(lines, m.filterOperatorView(width, bodyHeight)...)
	case modeColumns:
		lines = append(lines, m.columnsView(width, bodyHeight)...)
	default:
		lines = append(lines, m.logView(width, bodyHeight)...)
	}

	lines = append(lines, m.fit(width, m.inputLine()))
	lines = append(lines, m.fit(width, m.helpLine()))

	view := tea.NewView(strings.Join(lines, "\n"))
	view.AltScreen = true

	return view
}

// statusLine renders the top status bar.
func (m Model) statusLine(width int) string {
	state := "running"
	if m.paused {
		state = "paused"
	}
	if m.eof && len(m.pending) == 0 {
		state = "eof"
	}
	if m.err != nil {
		state = "error"
	}

	line := fmt.Sprintf(
		" loggo  profile=%s  parser=%s  state=%s  lines=%d  buffered=%d  visible=%d ",
		m.activeProfile.Name,
		m.activeParser.Name(),
		state,
		m.nextIndex,
		len(m.raw),
		len(m.visible),
	)

	if m.debug {
		line += fmt.Sprintf(" pending=%d version=%s commit=%s ", len(m.pending), m.version, m.commit)
	}

	badge := m.streamStateBadge()
	if width > 0 {
		gap := width - lipgloss.Width(line) - lipgloss.Width(badge)
		if gap < 1 {
			gap = 1
		}

		line += strings.Repeat(" ", gap) + badge
	} else {
		line += " " + badge
	}

	return statusStyle.Width(width).Render(line)
}

// activeLine renders the active search and filter state.
func (m Model) activeLine() string {
	search := "none"
	if m.search != "" {
		search = m.search
	}

	filters := "none"
	if len(m.include) > 0 {
		filters = strings.Join(m.include, ", ")
	}

	excludes := "none"
	if len(m.exclude) > 0 {
		excludes = strings.Join(m.exclude, ", ")
	}

	badges := m.badges()
	line := ""

	if len(badges) > 0 {
		line = strings.Join(badges, " ") + " "
	}

	line += fmt.Sprintf("search: %s   filters: %s   excludes: %s", search, filters, excludes)

	if len(m.hiddenFields) > 0 {
		line += fmt.Sprintf("   hidden fields: %d", len(m.hiddenFields))
	}
	if m.showHeaders {
		line += "   headers: on"
	} else {
		line += "   headers: off"
	}
	if m.filterContext > 0 && m.hasInteractiveFilters() {
		line += fmt.Sprintf("   context: %d", m.filterContext)
	}
	if m.horizontalOffset > 0 {
		line += fmt.Sprintf("   x:%d", m.horizontalOffset)
	}

	if m.err != nil {
		line += "   " + errorStyle.Render("error: "+m.err.Error())
	}

	return dimStyle.Render(line)
}

// badges renders visible badges for active search, filter, exclude, and column state.
func (m Model) badges() []string {
	badges := []string{}

	if strings.TrimSpace(m.search) != "" {
		badges = append(badges, badgeStyle.Render("SEARCH"))
	}
	if len(m.include) > 0 {
		badges = append(badges, badgeStyle.Render("FILTERED"))
	}
	if len(m.exclude) > 0 {
		badges = append(badges, badgeStyle.Render("HIDDEN"))
	}
	if len(m.hiddenFields) > 0 {
		badges = append(badges, badgeStyle.Render("COLUMNS"))
	}

	return badges
}

// logView renders the scrolling log viewport.
func (m Model) logView(width int, height int) []string {
	lines := make([]string, 0, height)
	bodyHeight := height

	if len(m.visible) == 0 {
		if header := m.headerLine(nil); header != "" && m.showHeaders {
			lines = append(lines, viewportLine(headerStyle.Render(header), m.horizontalOffset, width))
		}
		lines = append(lines, dimStyle.Render(" no visible log lines"))

		return padLines(lines, height)
	}

	start := m.selected - bodyHeight/2
	if m.follow {
		start = len(m.visible) - bodyHeight
	}

	start = max(0, start)
	end := min(len(m.visible), start+bodyHeight)
	fields := m.displayFields()

	if header := m.headerLine(fields); header != "" && m.showHeaders {
		lines = append(lines, viewportLine(headerStyle.Render(header), m.horizontalOffset, width))
		bodyHeight--
		if bodyHeight < 1 {
			return padLines(lines, height)
		}
		start = m.selected - bodyHeight/2
		if m.follow {
			start = len(m.visible) - bodyHeight
		}
		start = max(0, start)
		end = min(len(m.visible), start+bodyHeight)
	}

	for visibleIndex := start; visibleIndex < end; visibleIndex++ {
		entry := m.parsed[m.visible[visibleIndex]]
		selected := visibleIndex == m.selected
		matched := m.matchesSearch(entry)

		prefix := " "
		if selected {
			prefix = ">"
		}

		lines = append(lines, m.renderLogLine(width, prefix, entry, fields, selected, matched))
	}

	return padLines(lines, height)
}

// inputLine renders the command input line.
func (m Model) inputLine() string {
	switch m.mode {
	case modeSearch:
		return "/ " + m.input
	case modeFilterField:
		return "filter field> up/down select, enter apply, esc cancel"
	case modeFilterOperator:
		return "filter operator> up/down select, enter apply, esc cancel"
	case modeFilterValue:
		return fmt.Sprintf("filter %s %s> %s", m.filterField, m.filterOperator.Label, m.input)
	case modeExcludeField:
		return "exclude field> up/down select, enter apply, esc cancel"
	case modeExcludeOperator:
		return "exclude operator> up/down select, enter apply, esc cancel"
	case modeExcludeValue:
		return fmt.Sprintf("exclude %s %s> %s", m.filterField, m.filterOperator.Label, m.input)
	case modeColumns:
		return "columns> space toggle, enter apply, a show all, d profile default, esc cancel"
	case modeProfile:
		return "profile> up/down select, enter apply, esc cancel"
	case modeInspect:
		return "inspect> enter/esc close"
	case modeHelp:
		return "help> enter/esc close"
	default:
		return ""
	}
}

// helpLine renders the bottom help line.
func (m Model) helpLine() string {
	return dimStyle.Render("/ search  c clear  f filter  x exclude  h/l scroll  [/ ] context  v columns  H headers  F/X remove  r reset  p profile  ? help  q quit")
}

// streamStateBadge renders the current viewport follow state.
func (m Model) streamStateBadge() string {
	switch {
	case m.eof && len(m.pending) == 0:
		return eofStyle.Render("EOF")
	case m.follow && !m.paused:
		return liveStyle.Render("FOLLOWING")
	default:
		return frozenStyle.Render("FROZEN")
	}
}

// headerLine renders fixed column headers when using the default log row layout.
func (m Model) headerLine(fields []string) string {
	if m.activeProfile.Format != "" {
		return ""
	}

	parts := []string{}

	if !m.isHiddenField("timestamp") {
		parts = append(parts, formatColumn("TIMESTAMP", timestampColumnWidth))
	}
	if !m.isHiddenField("level") {
		parts = append(parts, formatColumn("LEVEL", levelColumnWidth))
	}
	for _, field := range fields {
		parts = append(parts, formatColumn(strings.ToUpper(field), fieldColumnWidth))
	}
	if !m.isHiddenField("message") {
		parts = append(parts, "MESSAGE")
	}

	return " " + strings.Join(parts, " ")
}

// renderEntry renders one parsed log entry.
func (m Model) renderEntry(entry logentry.Entry, fields []string, background string) string {
	if m.activeProfile.Format != "" {
		return m.renderFormat(entry)
	}

	parts := []string{}

	if entry.Timestamp != "" && !m.isHiddenField("timestamp") {
		parts = append(parts, m.renderCell(entry.Timestamp, m.activeProfile.Colors.Timestamp, timestampColumnWidth, background))
	}

	if entry.Level != "" && !m.isHiddenField("level") {
		color := m.activeProfile.Colors.Levels[strings.ToUpper(entry.Level)]
		parts = append(parts, m.renderCell(strings.ToUpper(entry.Level), color, levelColumnWidth, background))
	}

	for _, field := range fields {
		value, ok := entry.Get(field)
		if !ok || value == "" {
			parts = append(parts, m.renderCell("", "", fieldColumnWidth, background))
			continue
		}

		color := m.activeProfile.Colors.Fields[field]
		parts = append(parts, m.renderCell(field+"="+value, color, fieldColumnWidth, background))
	}

	message := entry.Message
	if message == "" {
		message = entry.Raw
	}

	if !m.isHiddenField("message") {
		parts = append(parts, m.renderCell(message, m.activeProfile.Colors.Message, 0, background))
	}

	return strings.Join(parts, " ")
}

// renderLogLine renders one visible log row with selection and search highlighting.
func (m Model) renderLogLine(width int, prefix string, entry logentry.Entry, fields []string, selected bool, matched bool) string {
	background := ""
	switch {
	case selected:
		background = "236"
	case matched:
		background = "11"
	}

	prefixStyle := lipgloss.NewStyle()
	if background != "" {
		prefixStyle = prefixStyle.Background(lipgloss.Color(background))
	}

	rendered := prefixStyle.Render(prefix) + m.renderEntry(entry, fields, background)

	return viewportLine(rendered, m.horizontalOffset, width)
}

// displayFields returns the configured fields that should currently be rendered.
func (m Model) displayFields() []string {
	fields := m.configuredFields()
	if m.activeProfile.FixedFields {
		return fields
	}

	return m.presentFields(fields)
}

// configuredFields returns configured non-core fields that are currently enabled.
func (m Model) configuredFields() []string {
	fields := []string{}

	for _, field := range m.activeProfile.Fields {
		if isCoreField(field) || m.isHiddenField(field) {
			continue
		}

		fields = append(fields, field)
	}

	return fields
}

// presentFields keeps field order stable while dropping columns that are empty
// across the currently visible dataset.
func (m Model) presentFields(fields []string) []string {
	if len(fields) == 0 || len(m.visible) == 0 {
		return fields
	}

	present := make([]string, 0, len(fields))

	for _, field := range fields {
		for _, index := range m.visible {
			value, ok := m.parsed[index].Get(field)
			if ok && value != "" {
				present = append(present, field)
				break
			}
		}
	}

	return present
}

// formatColumn pads one non-message column to a stable width.
func formatColumn(value string, width int) string {
	if width <= 0 {
		return value
	}

	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(value)
}

// renderCell renders one row cell with optional selection background.
func (m Model) renderCell(value string, color string, width int, background string) string {
	style := colorStyle(color)
	if width > 0 {
		style = style.Width(width).MaxWidth(width)
	}
	if background != "" {
		style = style.Background(lipgloss.Color(background))
	}

	return style.Render(value)
}

// renderFormat renders a configured format with field placeholders.
func (m Model) renderFormat(entry logentry.Entry) string {
	output := m.activeProfile.Format

	replacements := map[string]string{
		"raw":       entry.Raw,
		"timestamp": entry.Timestamp,
		"time":      entry.Timestamp,
		"level":     entry.Level,
		"message":   entry.Message,
		"msg":       entry.Message,
	}

	for key, value := range entry.Fields {
		replacements[key] = value
	}

	for key, value := range replacements {
		if m.isHiddenField(key) {
			value = ""
		}

		output = strings.ReplaceAll(output, "{"+key+"}", value)
	}

	return strings.TrimSpace(output)
}

// Update exists here only to ensure the tea.Model interface remains explicit.
var _ tea.Model = Model{}

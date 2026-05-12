package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gi8lino/loggo/internal/logentry"
)

// View renders the complete TUI.
func (m Model) View() string {
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
		m.fit(width, m.statusLine()),
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

	return strings.Join(lines, "\n")
}

// statusLine renders the top status bar.
func (m Model) statusLine() string {
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
		" loggo  profile=%s  parser=%s  follow=%t  state=%s  lines=%d  buffered=%d  visible=%d ",
		m.activeProfile.Name,
		m.activeParser.Name(),
		m.follow,
		state,
		m.nextIndex,
		len(m.raw),
		len(m.visible),
	)

	if m.debug {
		line += fmt.Sprintf(" pending=%d version=%s commit=%s ", len(m.pending), m.version, m.commit)
	}

	return statusStyle.Render(line)
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

	line += fmt.Sprintf("search: %s   filters: %s   hidden: %s", search, filters, excludes)

	if len(m.hiddenFields) > 0 {
		line += fmt.Sprintf("   hidden fields: %d", len(m.hiddenFields))
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

	if len(m.visible) == 0 {
		lines = append(lines, dimStyle.Render(" no visible log lines"))

		return padLines(lines, height)
	}

	start := m.selected - height/2
	if m.follow {
		start = len(m.visible) - height
	}

	start = max(0, start)
	end := min(len(m.visible), start+height)

	for visibleIndex := start; visibleIndex < end; visibleIndex++ {
		entry := m.parsed[m.visible[visibleIndex]]
		line := m.renderEntry(entry)

		if m.matchesSearch(entry) {
			line = matchStyle.Render(line)
		}

		if visibleIndex == m.selected {
			line = selectedStyle.Render(line)
		}

		prefix := " "
		if visibleIndex == m.selected {
			prefix = ">"
		}

		lines = append(lines, m.fit(width, prefix+line))
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
	return dimStyle.Render("/ search  c clear  f filter  x exclude  v columns  F/X remove  r reset  p profile  ? help  q quit")
}

// renderEntry renders one parsed log entry.
func (m Model) renderEntry(entry logentry.Entry) string {
	if m.activeProfile.Format != "" {
		return m.renderFormat(entry)
	}

	parts := []string{}

	if entry.Timestamp != "" && !m.isHiddenField("timestamp") {
		parts = append(parts, colorStyle(m.activeProfile.Colors.Timestamp).Render(entry.Timestamp))
	}

	if entry.Level != "" && !m.isHiddenField("level") {
		color := m.activeProfile.Colors.Levels[strings.ToUpper(entry.Level)]
		parts = append(parts, colorStyle(color).Render(padRight(strings.ToUpper(entry.Level), 5)))
	}

	for _, field := range m.activeProfile.Fields {
		if isCoreField(field) || m.isHiddenField(field) {
			continue
		}

		value, ok := entry.Get(field)
		if !ok || value == "" {
			continue
		}

		color := m.activeProfile.Colors.Fields[field]
		parts = append(parts, colorStyle(color).Render(field+"="+value))
	}

	message := entry.Message
	if message == "" {
		message = entry.Raw
	}

	if !m.isHiddenField("message") {
		parts = append(parts, colorStyle(m.activeProfile.Colors.Message).Render(message))
	}

	return strings.Join(parts, " ")
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

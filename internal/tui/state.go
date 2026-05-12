package tui

import (
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/gi8lino/loggo/internal/filter"
	"github.com/gi8lino/loggo/internal/logentry"
	logparser "github.com/gi8lino/loggo/internal/parser"
	"github.com/gi8lino/loggo/internal/profile"
)

var guidedOperators = []filterOperator{
	{Label: "contains", Token: "contains", NeedsValue: true},
	{Label: "equals", Token: "=", NeedsValue: true},
	{Label: "wildcard", Token: "wildcard", NeedsValue: true},
	{Label: "regex", Token: "regex", NeedsValue: true},
	{Label: "exists", Token: "exists", NeedsValue: false},
	{Label: "not equals", Token: "!=", NeedsValue: true},
	{Label: "greater than", Token: ">", NeedsValue: true},
	{Label: "greater or equal", Token: ">=", NeedsValue: true},
	{Label: "less than", Token: "<", NeedsValue: true},
	{Label: "less or equal", Token: "<=", NeedsValue: true},
	{Label: "in list", Token: "in", NeedsValue: true},
	{Label: "after", Token: "after", NeedsValue: true},
	{Label: "before", Token: "before", NeedsValue: true},
}

// startGuidedFilter starts the guided include or exclude filter flow.
func (m *Model) startGuidedFilter(exclude bool) {
	m.filterFieldOptions = m.buildFilterFields()
	m.filterFieldCursor = 0
	m.filterOperatorCursor = 0
	m.filterField = ""
	m.filterOperator = filterOperator{}
	m.input = ""

	if exclude {
		m.mode = modeExcludeField
		return
	}

	m.mode = modeFilterField
}

// cancelGuidedFilter clears guided filter state and returns to normal mode.
func (m *Model) cancelGuidedFilter() {
	m.filterField = ""
	m.filterOperator = filterOperator{}
	m.filterFieldOptions = nil
	m.input = ""
	m.mode = modeNormal
}

// applyGuidedFilter creates and applies a filter expression from the guided flow.
func (m *Model) applyGuidedFilter(value string) {
	expr := buildFilterExpression(m.filterField, m.filterOperator, value)

	if m.mode == modeExcludeValue || m.mode == modeExcludeOperator {
		next := append(append([]string{}, m.exclude...), expr)
		if err := m.setFilters(m.include, next); err != nil {
			m.err = err
			return
		}
	} else {
		next := append(append([]string{}, m.include...), expr)
		if err := m.setFilters(next, m.exclude); err != nil {
			m.err = err
			return
		}
	}

	m.input = ""
	m.filterField = ""
	m.filterOperator = filterOperator{}
	m.filterFieldOptions = nil
	m.mode = modeNormal
	m.err = nil
	m.rebuildVisible()
}

// buildFilterExpression builds a runtime filter expression.
func buildFilterExpression(field string, op filterOperator, value string) string {
	field = strings.TrimSpace(field)
	value = strings.TrimSpace(value)

	switch op.Token {
	case "exists":
		return field + " exists"
	case "=", "!=", ">", ">=", "<", "<=":
		return field + op.Token + value
	default:
		return field + " " + op.Token + " " + value
	}
}

// removeLastInclude removes the most recently added include filter.
func (m *Model) removeLastInclude() {
	if len(m.include) == 0 {
		return
	}

	m.include = m.include[:len(m.include)-1]
	m.err = nil
	_ = m.rebuildFilter()
	m.rebuildVisible()
}

// removeLastExclude removes the most recently added exclude filter.
func (m *Model) removeLastExclude() {
	if len(m.exclude) == 0 {
		return
	}

	m.exclude = m.exclude[:len(m.exclude)-1]
	m.err = nil
	_ = m.rebuildFilter()
	m.rebuildVisible()
}

// startColumnPicker starts the visible column picker.
func (m *Model) startColumnPicker() {
	m.columnFieldOptions = m.buildColumnFields()
	m.columnHiddenDraft = mapsClone(m.hiddenFields)
	m.columnFieldCursor = 0
	m.mode = modeColumns
}

// cancelColumnPicker discards column visibility changes.
func (m *Model) cancelColumnPicker() {
	m.columnFieldOptions = nil
	m.columnHiddenDraft = nil
	m.columnFieldCursor = 0
	m.mode = modeNormal
}

// applyColumnPicker applies column visibility changes.
func (m *Model) applyColumnPicker() {
	m.hiddenFields = mapsClone(m.columnHiddenDraft)
	m.columnFieldOptions = nil
	m.columnHiddenDraft = nil
	m.columnFieldCursor = 0
	m.mode = modeNormal
}

// toggleColumnDraft toggles the selected field in the draft hidden-field set.
func (m *Model) toggleColumnDraft() {
	if len(m.columnFieldOptions) == 0 {
		return
	}

	field := m.columnFieldOptions[m.columnFieldCursor]
	if _, ok := m.columnHiddenDraft[field]; ok {
		delete(m.columnHiddenDraft, field)
		return
	}

	m.columnHiddenDraft[field] = struct{}{}
}

// processPending parses and appends pending input lines during a frame.
func (m *Model) processPending(limit int) {
	if len(m.pending) == 0 {
		return
	}

	if limit <= 0 || limit > len(m.pending) {
		limit = len(m.pending)
	}

	lines := m.pending[:limit]
	m.pending = m.pending[limit:]

	for _, line := range lines {
		m.appendLineIncremental(line)
	}

	m.trimOverflow()
	m.syncSelectionAfterAppend()
}

// appendLineIncremental appends one line without rebuilding the full visible list.
func (m *Model) appendLineIncremental(line string) {
	raw := RawLine{
		Index:    m.nextIndex,
		Text:     line,
		Received: time.Now(),
	}

	m.nextIndex++

	entry := m.activeParser.Parse(line)
	index := len(m.parsed)

	m.raw = append(m.raw, raw)
	m.parsed = append(m.parsed, entry)

	if m.filterSet == nil || m.filterSet.Match(entry) {
		m.visible = append(m.visible, index)
	}
}

// trimOverflow trims old buffered lines and adjusts visible indexes.
func (m *Model) trimOverflow() {
	overflow := len(m.raw) - m.bufferSize
	if overflow <= 0 {
		return
	}

	m.raw = m.raw[overflow:]
	m.parsed = m.parsed[overflow:]

	nextVisible := m.visible[:0]
	for _, index := range m.visible {
		if index < overflow {
			continue
		}

		nextVisible = append(nextVisible, index-overflow)
	}

	m.visible = nextVisible
}

// syncSelectionAfterAppend keeps the selection valid after incoming logs.
func (m *Model) syncSelectionAfterAppend() {
	if len(m.visible) == 0 {
		m.selected = 0
		return
	}

	if m.follow && !m.paused {
		m.selected = len(m.visible) - 1
	}

	if m.selected >= len(m.visible) {
		m.selected = len(m.visible) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

// setFilters replaces runtime include and exclude filters.
func (m *Model) setFilters(include []string, exclude []string) error {
	oldInclude := m.include
	oldExclude := m.exclude

	m.include = include
	m.exclude = exclude

	if err := m.rebuildFilter(); err != nil {
		m.include = oldInclude
		m.exclude = oldExclude
		_ = m.rebuildFilter()

		return err
	}

	return nil
}

// rebuildFilter rebuilds the active filter set.
func (m *Model) rebuildFilter() error {
	set, err := filter.NewSet(m.activeProfile, m.include, m.exclude)
	if err != nil {
		return err
	}

	m.filterSet = set

	return nil
}

// switchProfile changes the active profile and reparses the raw buffer.
func (m *Model) switchProfile(name string) {
	next, ok := m.profiles[name]
	if !ok {
		return
	}

	next = profile.Normalize(name, next)

	parser, err := logparser.New(next)
	if err != nil {
		m.err = err
		return
	}

	m.activeProfile = next
	m.activeParser = parser
	m.profileCursor = m.activeProfileIndex()
	m.hiddenFields = fieldSet(next.HiddenFields)
	m.filterFieldOptions = nil
	m.columnFieldOptions = nil
	m.columnHiddenDraft = nil
	m.reparseAll()

	if err := m.rebuildFilter(); err != nil {
		m.err = err
	}

	m.rebuildVisible()
}

// reparseAll reparses every buffered raw line with the active profile.
func (m *Model) reparseAll() {
	m.parsed = m.parsed[:0]

	for _, raw := range m.raw {
		m.parsed = append(m.parsed, m.activeParser.Parse(raw.Text))
	}
}

// rebuildVisible recomputes visible indexes after global state changes.
func (m *Model) rebuildVisible() {
	m.visible = m.visible[:0]

	for index, entry := range m.parsed {
		if m.filterSet == nil || m.filterSet.Match(entry) {
			m.visible = append(m.visible, index)
		}
	}

	m.syncSelectionAfterAppend()
}

// moveSelection moves the selected visible entry.
func (m *Model) moveSelection(delta int) {
	if len(m.visible) == 0 {
		m.selected = 0
		return
	}

	m.follow = false
	m.selected = min(max(0, m.selected+delta), len(m.visible)-1)
}

// moveSearch moves to the next or previous visible search match.
func (m *Model) moveSearch(delta int) {
	if strings.TrimSpace(m.search) == "" || len(m.visible) == 0 {
		return
	}

	index := m.selected

	for range len(m.visible) {
		index += delta

		if index < 0 {
			index = len(m.visible) - 1
		}
		if index >= len(m.visible) {
			index = 0
		}

		entry := m.parsed[m.visible[index]]
		if m.matchesSearch(entry) {
			m.selected = index
			m.follow = false
			return
		}
	}
}

// matchesSearch reports whether entry matches the active search query.
func (m Model) matchesSearch(entry logentry.Entry) bool {
	search := strings.ToLower(strings.TrimSpace(m.search))
	if search == "" {
		return false
	}

	return strings.Contains(strings.ToLower(m.searchCorpus(entry)), search)
}

// searchCorpus builds searchable text from raw, core fields, and parsed fields.
func (m Model) searchCorpus(entry logentry.Entry) string {
	parts := []string{
		entry.Raw,
		entry.Timestamp,
		entry.Level,
		entry.Message,
	}

	keys := slices.Sorted(maps.Keys(entry.Fields))
	for _, key := range keys {
		parts = append(parts, key, entry.Fields[key])
	}

	return strings.Join(parts, " ")
}

// activeEntry returns the currently selected entry.
func (m Model) activeEntry() (logentry.Entry, bool) {
	if len(m.visible) == 0 || m.selected < 0 || m.selected >= len(m.visible) {
		return logentry.Entry{}, false
	}

	return m.parsed[m.visible[m.selected]], true
}

// activeProfileIndex returns the current active profile index.
func (m Model) activeProfileIndex() int {
	for index, name := range m.profileNames {
		if name == m.activeProfile.Name {
			return index
		}
	}

	return 0
}

// buildFilterFields returns a stable snapshot of selectable filter field names.
func (m Model) buildFilterFields() []string {
	seen := map[string]struct{}{}
	fields := []string{}
	discovered := map[string]struct{}{}

	add := func(field string) {
		field = strings.TrimSpace(field)
		if field == "" {
			return
		}
		if _, ok := seen[field]; ok {
			return
		}

		seen[field] = struct{}{}
		fields = append(fields, field)
	}

	add("raw")
	add("timestamp")
	add("time")
	add("level")
	add("message")
	add(m.activeProfile.TimestampField)
	add(m.activeProfile.LevelField)
	add(m.activeProfile.MessageField)

	for _, field := range m.activeProfile.Fields {
		add(field)
	}

	for _, entry := range m.parsed {
		for field := range entry.Fields {
			if _, ok := seen[field]; ok {
				continue
			}

			discovered[field] = struct{}{}
		}
	}

	for _, field := range slices.Sorted(maps.Keys(discovered)) {
		add(field)
	}

	return fields
}

// buildColumnFields returns a stable snapshot of displayable column names.
func (m Model) buildColumnFields() []string {
	seen := map[string]struct{}{}
	fields := []string{}
	discovered := map[string]struct{}{}

	add := func(field string) {
		field = strings.TrimSpace(field)
		if field == "" {
			return
		}
		if _, ok := seen[field]; ok {
			return
		}

		seen[field] = struct{}{}
		fields = append(fields, field)
	}

	add("timestamp")
	add("level")
	add("message")
	add(m.activeProfile.TimestampField)
	add(m.activeProfile.LevelField)
	add(m.activeProfile.MessageField)

	for _, field := range m.activeProfile.Fields {
		add(field)
	}

	for _, entry := range m.parsed {
		for field := range entry.Fields {
			if _, ok := seen[field]; ok {
				continue
			}

			discovered[field] = struct{}{}
		}
	}

	for _, field := range slices.Sorted(maps.Keys(discovered)) {
		add(field)
	}

	return fields
}

// isHiddenField reports whether a display field is hidden.
func (m Model) isHiddenField(field string) bool {
	field = strings.TrimSpace(field)
	if field == "" {
		return false
	}

	if _, ok := m.hiddenFields[field]; ok {
		return true
	}

	switch strings.ToLower(field) {
	case "timestamp", "time", "ts":
		return hasAnyField(m.hiddenFields, "timestamp", "time", "ts")
	case "level", "severity":
		return hasAnyField(m.hiddenFields, "level", "severity")
	case "message", "msg":
		return hasAnyField(m.hiddenFields, "message", "msg")
	default:
		return false
	}
}

// fieldSet creates a set from field names.
func fieldSet(fields []string) map[string]struct{} {
	set := map[string]struct{}{}

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		set[field] = struct{}{}
	}

	return set
}

// mapsClone returns a non-nil clone of a string set.
func mapsClone(input map[string]struct{}) map[string]struct{} {
	if input == nil {
		return map[string]struct{}{}
	}

	return maps.Clone(input)
}

// hasAnyField reports whether any field exists in a set.
func hasAnyField(set map[string]struct{}, fields ...string) bool {
	for _, field := range fields {
		if _, ok := set[field]; ok {
			return true
		}
	}

	return false
}

package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// handleKey routes key messages to the active mode handler.
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeSearch, modeFilterValue, modeExcludeValue:
		return m.handleInputKey(msg), nil
	case modeFilterField, modeExcludeField:
		return m.handleFilterFieldKey(msg), nil
	case modeFilterOperator, modeExcludeOperator:
		return m.handleFilterOperatorKey(msg), nil
	case modeColumns:
		return m.handleColumnsKey(msg), nil
	case modeProfile:
		return m.handleProfileKey(msg), nil
	case modeInspect, modeHelp:
		return m.handleOverlayKey(msg), nil
	default:
		return m.handleNormalKey(msg)
	}
}

// handleNormalKey handles keyboard input in normal mode.
func (m Model) handleNormalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "/":
		m.mode = modeSearch
		m.input = m.search
	case "f":
		m.startGuidedFilter(false)
	case "x":
		m.startGuidedFilter(true)
	case "F":
		m.removeLastInclude()
	case "X":
		m.removeLastExclude()
	case "c":
		m.search = ""
		m.input = ""
	case "v":
		m.startColumnPicker()
	case "p":
		m.mode = modeProfile
		m.profileCursor = m.activeProfileIndex()
	case "?":
		m.mode = modeHelp
	case "enter":
		if _, ok := m.activeEntry(); ok {
			m.mode = modeInspect
		}
	case "space":
		m.paused = !m.paused
		if m.paused {
			m.follow = false
		}
	case "a":
		m.paused = false
		m.follow = true
		m.selected = max(0, len(m.visible)-1)
	case "r":
		m.search = ""
		m.include = nil
		m.exclude = nil
		m.hiddenFields = fieldSet(m.activeProfile.HiddenFields)
		m.err = nil
		_ = m.rebuildFilter()
		m.rebuildVisible()
	case "up":
		m.moveSelection(-1)
	case "down":
		m.moveSelection(1)
	case "pgup":
		m.moveSelection(-10)
	case "pgdown":
		m.moveSelection(10)
	case "home":
		m.follow = false
		m.selected = 0
	case "end":
		m.follow = true
		m.paused = false
		m.selected = max(0, len(m.visible)-1)
	case "n":
		m.moveSearch(1)
	case "N":
		m.moveSearch(-1)
	}

	return m, nil
}

// handleInputKey handles keyboard input while editing a search or filter value.
func (m Model) handleInputKey(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		m.input = ""
	case "enter":
		m.applyInput()
	case "backspace", "ctrl+h":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
			if m.mode == modeSearch {
				m.search = m.input
			}
		}
	case "ctrl+u":
		m.input = ""
		if m.mode == modeSearch {
			m.search = ""
		}
	default:
		if len(msg.Text) > 0 {
			m.input += msg.Text
			if m.mode == modeSearch {
				m.search = m.input
			}
		}
	}

	return m
}

// handleFilterFieldKey handles field selection for guided filters.
func (m Model) handleFilterFieldKey(msg tea.KeyPressMsg) Model {
	if len(m.filterFieldOptions) == 0 {
		m.filterFieldOptions = m.buildFilterFields()
	}

	switch msg.String() {
	case "esc":
		m.cancelGuidedFilter()
	case "up":
		m.filterFieldCursor = max(0, m.filterFieldCursor-1)
	case "down":
		m.filterFieldCursor = min(len(m.filterFieldOptions)-1, m.filterFieldCursor+1)
	case "enter":
		if len(m.filterFieldOptions) == 0 {
			return m
		}

		m.filterField = m.filterFieldOptions[m.filterFieldCursor]
		m.filterOperatorCursor = 0

		if m.mode == modeExcludeField {
			m.mode = modeExcludeOperator
		} else {
			m.mode = modeFilterOperator
		}
	}

	return m
}

// handleFilterOperatorKey handles operator selection for guided filters.
func (m Model) handleFilterOperatorKey(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "esc":
		m.cancelGuidedFilter()
	case "up":
		m.filterOperatorCursor = max(0, m.filterOperatorCursor-1)
	case "down":
		m.filterOperatorCursor = min(len(guidedOperators)-1, m.filterOperatorCursor+1)
	case "enter":
		m.filterOperator = guidedOperators[m.filterOperatorCursor]

		if !m.filterOperator.NeedsValue {
			m.applyGuidedFilter("")
			return m
		}

		m.input = ""

		if m.mode == modeExcludeOperator {
			m.mode = modeExcludeValue
		} else {
			m.mode = modeFilterValue
		}
	}

	return m
}

// handleColumnsKey handles keyboard input in the column visibility picker.
func (m Model) handleColumnsKey(msg tea.KeyPressMsg) Model {
	if len(m.columnFieldOptions) == 0 {
		m.columnFieldOptions = m.buildColumnFields()
	}
	if m.columnHiddenDraft == nil {
		m.columnHiddenDraft = mapsClone(m.hiddenFields)
	}

	switch msg.String() {
	case "esc":
		m.cancelColumnPicker()
	case "enter":
		m.applyColumnPicker()
	case "up":
		m.columnFieldCursor = max(0, m.columnFieldCursor-1)
	case "down":
		m.columnFieldCursor = min(len(m.columnFieldOptions)-1, m.columnFieldCursor+1)
	case "space":
		m.toggleColumnDraft()
	case "a":
		m.columnHiddenDraft = map[string]struct{}{}
	case "d":
		m.columnHiddenDraft = fieldSet(m.activeProfile.HiddenFields)
	}

	return m
}

// handleProfileKey handles keyboard input in profile picker mode.
func (m Model) handleProfileKey(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
	case "up":
		m.profileCursor = max(0, m.profileCursor-1)
	case "down":
		m.profileCursor = min(len(m.profileNames)-1, m.profileCursor+1)
	case "enter":
		if m.profileCursor >= 0 && m.profileCursor < len(m.profileNames) {
			m.switchProfile(m.profileNames[m.profileCursor])
		}

		m.mode = modeNormal
	}

	return m
}

// handleOverlayKey handles keyboard input in help and inspect modes.
func (m Model) handleOverlayKey(msg tea.KeyPressMsg) Model {
	switch msg.String() {
	case "esc", "enter":
		m.mode = modeNormal
	case "q", "ctrl+c":
		m.mode = modeNormal
	}

	return m
}

// applyInput applies the current command-line input.
func (m *Model) applyInput() {
	value := strings.TrimSpace(m.input)

	switch m.mode {
	case modeSearch:
		m.search = value
	case modeFilterValue, modeExcludeValue:
		if value != "" {
			m.applyGuidedFilter(value)
			return
		}
	}

	m.input = ""
	m.mode = modeNormal
	m.err = nil
}

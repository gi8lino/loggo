package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleKeyStopsStreamBeforeQuit(t *testing.T) {
	stopped := false
	model := Model{
		stopStream: func() {
			stopped = true
		},
	}

	_, cmd := model.handleKey(tea.KeyPressMsg{Code: 'q', Text: "q"})

	assert.True(t, stopped)
	require.NotNil(t, cmd)
}

func TestHandleKeyStopsStreamOnCtrlCInSearchMode(t *testing.T) {
	stopped := false
	model := Model{
		mode: modeSearch,
		stopStream: func() {
			stopped = true
		},
	}

	_, cmd := model.handleKey(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	assert.True(t, stopped)
	require.NotNil(t, cmd)
}

func TestHandleKeyTogglesHeaders(t *testing.T) {
	model := Model{showHeaders: true}

	next, _ := model.handleKey(tea.KeyPressMsg{Code: 'H', Text: "H"})
	updated := next.(Model)
	assert.False(t, updated.showHeaders)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: 'H', Text: "H"})
	updated = next.(Model)
	assert.True(t, updated.showHeaders)
}

func TestStructuredSearchUsesFieldAwareMatcher(t *testing.T) {
	model := Model{}
	model.setSearchQuery("level:ERROR")

	entry := logentry.New("raw")
	entry.Level = "ERROR"

	assert.True(t, model.matchesSearch(entry))

	model.setSearchQuery("service = billing-api and level = ERROR")
	entry.Fields["service"] = "billing-api"

	assert.True(t, model.matchesSearch(entry))
}

func TestHandleKeySupportsVimNavigation(t *testing.T) {
	model := Model{
		visible:  []int{0, 1, 2},
		selected: 1,
	}

	next, _ := model.handleKey(tea.KeyPressMsg{Code: 'j', Text: "j"})
	updated := next.(Model)
	assert.Equal(t, 2, updated.selected)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: 'k', Text: "k"})
	updated = next.(Model)
	assert.Equal(t, 1, updated.selected)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: 'g', Text: "g"})
	updated = next.(Model)
	assert.True(t, updated.vimGotoPending)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: 'g', Text: "g"})
	updated = next.(Model)
	assert.Equal(t, 0, updated.selected)
	assert.False(t, updated.vimGotoPending)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: 'G', Text: "G"})
	updated = next.(Model)
	assert.Equal(t, 2, updated.selected)
}

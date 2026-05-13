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

func TestHandleKeyAdjustsFilterContext(t *testing.T) {
	model := Model{
		include: []string{"level = ERROR"},
		parsed: []logentry.Entry{
			{Level: "INFO", Fields: map[string]string{}},
			{Level: "ERROR", Fields: map[string]string{}},
			{Level: "INFO", Fields: map[string]string{}},
		},
	}

	require.NoError(t, model.rebuildFilter())
	model.rebuildVisible()

	next, _ := model.handleKey(tea.KeyPressMsg{Code: ']', Text: "]"})
	updated := next.(Model)
	assert.Equal(t, 1, updated.filterContext)
	assert.Equal(t, []int{0, 1, 2}, updated.visible)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: '[', Text: "["})
	updated = next.(Model)
	assert.Equal(t, 0, updated.filterContext)
	assert.Equal(t, []int{1}, updated.visible)
}

func TestFilterContextKeepsSelectionAnchoredToSameRawLine(t *testing.T) {
	model := Model{
		include: []string{"level = ERROR"},
		parsed: []logentry.Entry{
			{Level: "INFO", Fields: map[string]string{}},
			{Level: "ERROR", Fields: map[string]string{}},
			{Level: "INFO", Fields: map[string]string{}},
			{Level: "ERROR", Fields: map[string]string{}},
			{Level: "INFO", Fields: map[string]string{}},
		},
	}

	require.NoError(t, model.rebuildFilter())
	model.rebuildVisible()
	model.selected = 1

	next, _ := model.handleKey(tea.KeyPressMsg{Code: ']', Text: "]"})
	updated := next.(Model)
	require.Equal(t, []int{0, 1, 2, 3, 4}, updated.visible)
	assert.Equal(t, 3, updated.visible[updated.selected])

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: '[', Text: "["})
	updated = next.(Model)
	require.Equal(t, []int{1, 3}, updated.visible)
	assert.Equal(t, 3, updated.visible[updated.selected])
}

func TestStreamStateBadgeReflectsFollowMode(t *testing.T) {
	model := Model{follow: true}
	assert.Contains(t, model.streamStateBadge(), "FOLLOWING")

	model.follow = false
	assert.Contains(t, model.streamStateBadge(), "FROZEN")

	model.eof = true
	assert.Contains(t, model.streamStateBadge(), "EOF")
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

func TestHandleKeySupportsHorizontalScrolling(t *testing.T) {
	model := Model{}

	next, _ := model.handleKey(tea.KeyPressMsg{Code: 'l', Text: "l"})
	updated := next.(Model)
	assert.Equal(t, 8, updated.horizontalOffset)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: 'h', Text: "h"})
	updated = next.(Model)
	assert.Equal(t, 0, updated.horizontalOffset)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: tea.KeyRight})
	updated = next.(Model)
	assert.Equal(t, 8, updated.horizontalOffset)

	next, _ = updated.handleKey(tea.KeyPressMsg{Code: tea.KeyLeft})
	updated = next.(Model)
	assert.Equal(t, 0, updated.horizontalOffset)
}

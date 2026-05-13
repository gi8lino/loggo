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

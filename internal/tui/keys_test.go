package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
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

func TestApplyInputJumpsToFirstSearchMatch(t *testing.T) {
	model := Model{
		mode:  modeSearch,
		input: "account is locked",
		parsed: []logentry.Entry{
			logentry.New("healthy"),
			logentry.New("ORA-28000: The account is locked."),
			logentry.New("healthy again"),
		},
		visible: []int{0, 1, 2},
	}

	model.applyInput()

	assert.Equal(t, "account is locked", model.search)
	assert.Equal(t, 1, model.selected)
	assert.False(t, model.follow)
	assert.Equal(t, modeNormal, model.mode)
}

func TestApplyInputRepeatingSearchMovesToNextMatch(t *testing.T) {
	model := Model{
		mode:     modeSearch,
		search:   "account is locked",
		input:    "account is locked",
		selected: 1,
		parsed: []logentry.Entry{
			logentry.New("healthy"),
			logentry.New("ORA-28000: The account is locked."),
			logentry.New("healthy again"),
			logentry.New("The account is locked for user"),
		},
		visible: []int{0, 1, 2, 3},
	}

	model.applyInput()

	assert.Equal(t, 3, model.selected)
	assert.Equal(t, modeNormal, model.mode)
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

func TestApplyInputExportsCurrentProfileSnapshot(t *testing.T) {
	var (
		savedName    string
		savedProfile profile.Profile
	)

	model := Model{
		mode:  modeExportProfile,
		input: "incident-view",
		activeProfile: profile.Normalize("json", profile.Profile{
			Parser: profile.ParserJSON,
			Fields: []string{"service", "status"},
			Filters: profile.Filters{
				Exclude: []profile.Rule{{Field: "path", Op: "equals", Value: "/health"}},
			},
		}),
		include:      []string{"status >= 500"},
		exclude:      []string{"user_agent wildcard *kube-probe*"},
		hiddenFields: fieldSet([]string{"status"}),
		parsed: []logentry.Entry{
			{Fields: map[string]string{"service": "billing", "status": "503", "method": "GET"}},
		},
		visible: []int{0},
		exportProfile: func(name string, p profile.Profile) (string, error) {
			savedName = name
			savedProfile = p
			return "/tmp/.loggo.yaml", nil
		},
	}

	model.applyInput()

	assert.Equal(t, modeNormal, model.mode)
	assert.Equal(t, "incident-view", savedName)
	assert.Equal(t, profile.ParserJSON, savedProfile.Parser)
	assert.Equal(t, []string{"service", "method"}, savedProfile.Fields)
	assert.Equal(t, []string{"status"}, savedProfile.HiddenFields)
	require.Len(t, savedProfile.Filters.Include, 1)
	assert.Equal(t, "status >= 500", savedProfile.Filters.Include[0].Expr)
	require.Len(t, savedProfile.Filters.Exclude, 2)
	assert.Equal(t, "/health", savedProfile.Filters.Exclude[0].Value)
	assert.Equal(t, "user_agent wildcard *kube-probe*", savedProfile.Filters.Exclude[1].Expr)
	assert.Contains(t, model.notice, "saved profile incident-view")
}

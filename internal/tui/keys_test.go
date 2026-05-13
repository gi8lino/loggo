package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gi8lino/loggo/internal/logentry"
)

func TestHandleKeyStopsStreamBeforeQuit(t *testing.T) {
	stopped := false
	model := Model{
		stopStream: func() {
			stopped = true
		},
	}

	_, cmd := model.handleKey(tea.KeyPressMsg{Code: 'q', Text: "q"})

	if !stopped {
		t.Fatalf("expected quit to stop the ingest stream")
	}
	if cmd == nil {
		t.Fatalf("expected quit command to be returned")
	}
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

	if !stopped {
		t.Fatalf("expected ctrl+c to stop the ingest stream")
	}
	if cmd == nil {
		t.Fatalf("expected quit command to be returned")
	}
}

func TestStructuredSearchUsesFieldAwareMatcher(t *testing.T) {
	model := Model{}
	model.setSearchQuery("level:ERROR")

	entry := logentry.New("raw")
	entry.Level = "ERROR"

	if !model.matchesSearch(entry) {
		t.Fatalf("expected field-aware search to match level alias")
	}

	model.setSearchQuery("service = billing-api and level = ERROR")
	entry.Fields["service"] = "billing-api"

	if !model.matchesSearch(entry) {
		t.Fatalf("expected structured search to use filter parser")
	}
}

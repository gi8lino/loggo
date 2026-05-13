package tui

import (
	"strings"
	"testing"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
	"github.com/stretchr/testify/assert"
)

func TestDisplayFieldsDropsColumnsMissingFromVisibleData(t *testing.T) {
	model := Model{
		activeProfile: profile.Profile{
			Fields: []string{"service", "trace_id", "status"},
		},
		parsed: []logentry.Entry{
			{Fields: map[string]string{"service": "billing", "logger": "api"}},
			{Fields: map[string]string{"service": "orders"}},
		},
		visible: []int{0, 1},
	}

	assert.Equal(t, []string{"service", "logger"}, model.displayFields())
}

func TestDisplayFieldsKeepsConfiguredOrderAcrossVisibleDataset(t *testing.T) {
	model := Model{
		activeProfile: profile.Profile{
			Fields: []string{"service", "trace_id", "status"},
		},
		parsed: []logentry.Entry{
			{Fields: map[string]string{"service": "billing"}},
			{Fields: map[string]string{"status": "500"}},
			{Fields: map[string]string{"trace_id": "abc"}},
		},
		visible: []int{0, 1, 2},
	}

	assert.Equal(t, []string{"service", "trace_id", "status"}, model.displayFields())
}

func TestDisplayFieldsAppendsDiscoveredFieldsAfterConfiguredOnes(t *testing.T) {
	model := Model{
		activeProfile: profile.Profile{
			Fields: []string{"service", "trace_id"},
		},
		parsed: []logentry.Entry{
			{Fields: map[string]string{"status": "500", "method": "GET", "service": "billing"}},
		},
		visible: []int{0},
	}

	assert.Equal(t, []string{"service", "method", "status"}, model.displayFields())
}

func TestDisplayFieldsRespectsFixedProfileColumns(t *testing.T) {
	model := Model{
		activeProfile: profile.Profile{
			Fields:      []string{"service", "trace_id", "status"},
			FixedFields: true,
		},
		parsed: []logentry.Entry{
			{Fields: map[string]string{"service": "billing"}},
		},
		visible: []int{0},
	}

	assert.Equal(t, []string{"service", "trace_id", "status"}, model.displayFields())
}

func TestRenderCellKeepsStructuredRowsSingleLine(t *testing.T) {
	model := Model{}

	rendered := model.renderCell("hostName=avaloq-aws-0", "", fieldColumnWidth, "")

	assert.False(t, strings.Contains(rendered, "\n"))
}

func TestRenderEntryUsesColumnValueWithoutFieldPrefix(t *testing.T) {
	model := Model{
		activeProfile: profile.Profile{
			Colors: profile.Colors{
				Levels: map[string]string{},
				Fields: map[string]string{},
			},
		},
	}
	entry := logentry.Entry{
		Fields: map[string]string{"sequence": "2514043"},
	}

	rendered := model.renderEntry(entry, []string{"sequence"}, "")

	assert.Contains(t, rendered, "2514043")
	assert.NotContains(t, rendered, "sequence=2514043")
}

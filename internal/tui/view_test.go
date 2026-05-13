package tui

import (
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
			{Fields: map[string]string{"service": "billing"}},
			{Fields: map[string]string{"service": "orders"}},
		},
		visible: []int{0, 1},
	}

	assert.Equal(t, []string{"service"}, model.displayFields())
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

package tui

import (
	"strings"
	"testing"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectViewWrapsRawPayload(t *testing.T) {
	entry := logentry.New("this is a very long raw payload that should wrap in inspect view")

	model := Model{
		parsed:   []logentry.Entry{entry},
		visible:  []int{0},
		selected: 0,
	}

	lines := model.inspectView(20, 20)

	rawIndex := -1
	for index, line := range lines {
		if strings.TrimSpace(line) == "raw" {
			rawIndex = index
			break
		}
	}

	require.NotEqual(t, -1, rawIndex)
	require.GreaterOrEqual(t, len(lines), rawIndex+3)
	assert.Equal(t, "---", strings.TrimSpace(lines[rawIndex+1]))

	rawLines := []string{}
	for _, line := range lines[rawIndex+2:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		rawLines = append(rawLines, trimmed)
	}

	require.GreaterOrEqual(t, len(rawLines), 2)
	assert.Contains(t, rawLines[0], "this is a very long")
	assert.Contains(t, strings.Join(rawLines, " "), "raw payload that should wrap in inspect view")
}

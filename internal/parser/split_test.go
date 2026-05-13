package parser

import (
	"testing"

	"github.com/gi8lino/loggo/internal/profile"
	"github.com/stretchr/testify/assert"
)

func TestSplitParserUsesRemainderForLastField(t *testing.T) {
	parser := NewSplit(profile.Normalize("pipe", profile.Profile{
		Parser: profile.ParserSplit,
		Split: profile.SplitConfig{
			Delimiter: "|",
			Fields:    []string{"time", "level", "message"},
		},
		TimestampField: "time",
		LevelField:     "level",
		MessageField:   "message",
	}))

	entry := parser.Parse(`2026-05-13T10:15:30Z|ERROR|first|second|third`)

	assert.Equal(t, "first|second|third", entry.Fields["message"])
}

func TestSplitParserHandlesQuotedSpaceDelimitedFields(t *testing.T) {
	parser := NewSplit(profile.Normalize("space", profile.Profile{
		Parser: profile.ParserSplit,
		Split: profile.SplitConfig{
			Delimiter: " ",
			Fields:    []string{"level", "component", "message"},
		},
		LevelField:   "level",
		MessageField: "message",
	}))

	entry := parser.Parse(`INFO "auth service" login failed for bob`)

	assert.Equal(t, "auth service", entry.Fields["component"])
	assert.Equal(t, "login failed for bob", entry.Fields["message"])
}

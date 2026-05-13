package parser

import (
	"testing"

	"github.com/gi8lino/loggo/internal/profile"
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

	if got := entry.Fields["message"]; got != "first|second|third" {
		t.Fatalf("expected remainder field to keep delimiter tail, got %q", got)
	}
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

	if got := entry.Fields["component"]; got != "auth service" {
		t.Fatalf("expected quoted field to be unquoted, got %q", got)
	}

	if got := entry.Fields["message"]; got != "login failed for bob" {
		t.Fatalf("expected final field to keep remainder, got %q", got)
	}
}

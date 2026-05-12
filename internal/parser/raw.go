package parser

import (
	"strings"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// Raw parses lines without structure.
type Raw struct{}

// NewRaw creates a raw parser.
func NewRaw() Parser {
	return Raw{}
}

// Name returns the parser name.
func (p Raw) Name() string {
	return profile.ParserRaw
}

// Parse creates a raw entry with basic level detection.
func (p Raw) Parse(line string) logentry.Entry {
	entry := logentry.New(line)
	entry.Parser = p.Name()
	entry.Level = detectLevel(line)
	entry.Message = line

	return entry
}

// detectLevel detects a level from a raw line.
func detectLevel(line string) string {
	upper := strings.ToUpper(line)

	for _, level := range []string{"PANIC", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"} {
		if strings.Contains(upper, level) {
			return level
		}
	}

	return ""
}

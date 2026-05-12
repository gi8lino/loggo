package parser

import (
	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// Auto tries structured parsers before falling back to raw parsing.
type Auto struct {
	parsers []Parser
	raw     Parser
}

// NewAuto creates an auto-detecting parser.
func NewAuto(p profile.Profile) (Parser, error) {
	parsers := []Parser{
		NewJSON(p),
		NewLogfmt(p),
	}

	if p.Regex != "" {
		regexParser, err := NewRegex(p)
		if err != nil {
			return nil, err
		}

		parsers = append(parsers, regexParser)
	}

	return Auto{
		parsers: parsers,
		raw:     NewRaw(),
	}, nil
}

// Name returns the parser name.
func (p Auto) Name() string {
	return profile.ParserAuto
}

// Parse parses a line using the first parser that succeeds.
func (p Auto) Parse(line string) logentry.Entry {
	for _, parser := range p.parsers {
		entry := parser.Parse(line)
		if entry.Parsed {
			return entry
		}
	}

	return p.raw.Parse(line)
}

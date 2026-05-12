package parser

import (
	"fmt"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// Parser parses one raw log line into a normalized entry.
type Parser interface {
	Name() string
	Parse(line string) logentry.Entry
}

// New creates a parser for a profile.
func New(p profile.Profile) (Parser, error) {
	switch p.Parser {
	case profile.ParserAuto:
		return NewAuto(p)
	case profile.ParserJSON:
		return NewJSON(p), nil
	case profile.ParserLogfmt:
		return NewLogfmt(p), nil
	case profile.ParserRegex:
		return NewRegex(p)
	case profile.ParserSplit:
		return NewSplit(p), nil
	case profile.ParserRaw:
		return NewRaw(), nil
	default:
		return nil, fmt.Errorf("unknown parser %q", p.Parser)
	}
}

package parser

import (
	"fmt"
	"regexp"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// Regex parses log lines with named capture groups.
type Regex struct {
	profile profile.Profile
	regex   *regexp.Regexp
	names   []string
}

// NewRegex creates a regex parser.
func NewRegex(p profile.Profile) (Parser, error) {
	if p.Regex == "" {
		return nil, fmt.Errorf("regex parser requires regex pattern")
	}

	compiled, err := regexp.Compile(p.Regex)
	if err != nil {
		return nil, fmt.Errorf("compile regex parser: %w", err)
	}

	return Regex{
		profile: p,
		regex:   compiled,
		names:   compiled.SubexpNames(),
	}, nil
}

// Name returns the parser name.
func (p Regex) Name() string {
	return profile.ParserRegex
}

// Parse parses one regex-matched log line.
func (p Regex) Parse(line string) logentry.Entry {
	matches := p.regex.FindStringSubmatch(line)
	if matches == nil {
		return logentry.Entry{}
	}

	entry := logentry.New(line)
	entry.Parser = p.Name()
	entry.Parsed = true

	for index, value := range matches {
		if index == 0 || index >= len(p.names) {
			continue
		}

		name := p.names[index]
		if name == "" {
			continue
		}

		entry.Fields[name] = value
	}

	applyCommonFields(&entry, p.profile)

	return entry
}

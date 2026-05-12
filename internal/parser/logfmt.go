package parser

import (
	"strconv"
	"strings"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// Logfmt parses logfmt-style key-value log lines.
type Logfmt struct {
	profile profile.Profile
}

// NewLogfmt creates a logfmt parser.
func NewLogfmt(p profile.Profile) Parser {
	return Logfmt{profile: p}
}

// Name returns the parser name.
func (p Logfmt) Name() string {
	return profile.ParserLogfmt
}

// Parse parses one logfmt log line.
func (p Logfmt) Parse(line string) logentry.Entry {
	fields := scanLogfmt(line)
	if len(fields) == 0 {
		return logentry.Entry{}
	}

	entry := logentry.New(line)
	entry.Parser = p.Name()
	entry.Parsed = true
	entry.Fields = fields

	applyCommonFields(&entry, p.profile)

	return entry
}

// scanLogfmt scans key=value pairs with basic quoted value support.
func scanLogfmt(line string) map[string]string {
	fields := map[string]string{}
	input := strings.TrimSpace(line)

	for len(input) > 0 {
		input = strings.TrimLeft(input, " \t")
		if input == "" {
			break
		}

		eq := strings.IndexByte(input, '=')
		if eq <= 0 {
			next := strings.IndexAny(input, " \t")
			if next < 0 {
				break
			}

			input = input[next+1:]

			continue
		}

		key := strings.TrimSpace(input[:eq])
		rest := input[eq+1:]

		if key == "" {
			break
		}

		var value string
		var consumed int

		if strings.HasPrefix(rest, `"`) {
			value, consumed = readQuoted(rest)
		} else {
			value, consumed = readBare(rest)
		}

		if consumed <= 0 {
			break
		}

		fields[key] = value
		input = rest[consumed:]
	}

	return fields
}

// readQuoted reads a quoted logfmt value.
func readQuoted(input string) (string, int) {
	escaped := false

	for i := 1; i < len(input); i++ {
		switch {
		case escaped:
			escaped = false
		case input[i] == '\\':
			escaped = true
		case input[i] == '"':
			raw := input[:i+1]
			value, err := strconv.Unquote(raw)
			if err != nil {
				return strings.Trim(raw, `"`), i + 1
			}

			return value, i + 1
		}
	}

	return strings.Trim(input, `"`), len(input)
}

// readBare reads an unquoted logfmt value.
func readBare(input string) (string, int) {
	end := strings.IndexAny(input, " \t")
	if end < 0 {
		return input, len(input)
	}

	return input[:end], end
}

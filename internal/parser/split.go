package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// Split parses delimiter-separated log lines.
type Split struct {
	profile profile.Profile
}

// NewSplit creates a split parser.
func NewSplit(p profile.Profile) Parser {
	return Split{profile: p}
}

// Name returns the parser name.
func (p Split) Name() string {
	return profile.ParserSplit
}

// Parse parses one delimiter-separated log line.
func (p Split) Parse(line string) logentry.Entry {
	entry := logentry.New(line)
	entry.Parser = p.Name()
	entry.Parsed = true

	for index, part := range p.splitLine(line) {
		p.addField(&entry, index, part)
	}

	applyCommonFields(&entry, p.profile)

	if entry.Message == "" {
		entry.Message = line
	}

	return entry
}

// splitLine splits a line according to the configured delimiter and fields.
func (p Split) splitLine(line string) []string {
	fieldCount := len(p.profile.Split.Fields)
	if fieldCount == 0 {
		return splitAllFields(line, p.profile.Split.Delimiter)
	}

	if fieldCount == 1 {
		return []string{strings.TrimSpace(line)}
	}

	if p.profile.Split.Delimiter == " " {
		return splitSpaceFields(line, fieldCount)
	}

	parts := strings.SplitN(line, p.profile.Split.Delimiter, fieldCount)
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}

	return parts
}

// splitAllFields splits a line into all fields without using a remainder field.
func splitAllFields(line string, delimiter string) []string {
	if delimiter == " " {
		parts := []string{}
		rest := strings.TrimSpace(line)

		for rest != "" {
			part, consumed := readSpaceToken(rest)
			if consumed <= 0 {
				break
			}

			parts = append(parts, part)
			rest = strings.TrimSpace(rest[consumed:])
		}

		return parts
	}

	parts := []string{}
	for part := range strings.SplitSeq(line, delimiter) {
		parts = append(parts, strings.TrimSpace(part))
	}

	return parts
}

// splitSpaceFields splits a space-delimited line and keeps the remainder in the final field.
func splitSpaceFields(line string, fieldCount int) []string {
	parts := make([]string, 0, fieldCount)
	rest := strings.TrimSpace(line)

	for len(parts) < fieldCount-1 && rest != "" {
		part, consumed := readSpaceToken(rest)
		if consumed <= 0 {
			break
		}

		parts = append(parts, part)
		rest = strings.TrimSpace(rest[consumed:])
	}

	if rest != "" {
		parts = append(parts, strings.TrimSpace(rest))
	}

	return parts
}

// readSpaceToken reads one possibly quoted token from a space-delimited string.
func readSpaceToken(input string) (string, int) {
	input = strings.TrimLeftFunc(input, unicode.IsSpace)
	if input == "" {
		return "", 0
	}

	if input[0] == '"' || input[0] == '\'' {
		quote := input[0]
		for index := 1; index < len(input); index++ {
			if input[index] != quote || input[index-1] == '\\' {
				continue
			}

			raw := input[:index+1]
			value, err := strconv.Unquote(raw)
			if err != nil {
				return strings.Trim(raw, string(quote)), index + 1
			}

			return value, index + 1
		}

		return strings.Trim(input, string(quote)), len(input)
	}

	for index, r := range input {
		if unicode.IsSpace(r) {
			return input[:index], index
		}
	}

	return input, len(input)
}

// addField adds one positional split field to the entry.
func (p Split) addField(entry *logentry.Entry, index int, part string) {
	value := strings.TrimSpace(part)
	name := fmt.Sprintf("col%d", index)

	if index < len(p.profile.Split.Fields) && strings.TrimSpace(p.profile.Split.Fields[index]) != "" {
		name = strings.TrimSpace(p.profile.Split.Fields[index])
	}

	entry.Fields[name] = value
}

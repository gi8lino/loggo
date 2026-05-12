package parser

import (
	"fmt"
	"strings"

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

	index := 0
	if p.profile.Split.Delimiter == " " {
		for part := range strings.FieldsSeq(line) {
			p.addField(&entry, index, part)
			index++
		}
	} else {
		for part := range strings.SplitSeq(line, p.profile.Split.Delimiter) {
			p.addField(&entry, index, part)
			index++
		}
	}

	applyCommonFields(&entry, p.profile)

	if entry.Message == "" {
		entry.Message = line
	}

	return entry
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

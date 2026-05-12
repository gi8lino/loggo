package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// JSON parses JSON log lines.
type JSON struct {
	profile profile.Profile
}

// NewJSON creates a JSON parser.
func NewJSON(p profile.Profile) Parser {
	return JSON{profile: p}
}

// Name returns the parser name.
func (p JSON) Name() string {
	return profile.ParserJSON
}

// Parse parses one JSON log line.
func (p JSON) Parse(line string) logentry.Entry {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "{") {
		return logentry.Entry{}
	}

	values := map[string]any{}
	if err := json.Unmarshal([]byte(trimmed), &values); err != nil {
		return logentry.Entry{}
	}

	entry := logentry.New(line)
	entry.Parser = p.Name()
	entry.Parsed = true

	for key, value := range values {
		entry.Fields[key] = stringify(value)
	}

	applyCommonFields(&entry, p.profile)

	return entry
}

// stringify converts a parsed JSON value to a string.
func stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		content, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}

		return string(content)
	}
}

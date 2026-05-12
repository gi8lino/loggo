package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/loggo/internal/logentry"
)

// timeMatcher compares parsed entry timestamps.
type timeMatcher struct {
	op       string
	expected time.Time
	raw      string
}

// newTimeMatcher creates a timestamp matcher.
func newTimeMatcher(field string, op string, value string, raw string) (Matcher, error) {
	if !isTimeField(field) && field != "" {
		return nil, fmt.Errorf("time operators only support timestamp fields")
	}

	op = normalizeTimeOp(op)

	parsed, ok := parseFilterTime(value)
	if !ok {
		return nil, fmt.Errorf("invalid time filter %q", raw)
	}

	return timeMatcher{
		op:       op,
		expected: parsed,
		raw:      raw,
	}, nil
}

// Match reports whether the entry timestamp matches.
func (m timeMatcher) Match(entry logentry.Entry) bool {
	if !entry.HasTime {
		return false
	}

	switch m.op {
	case "gt":
		return entry.Time.After(m.expected)
	case "gte":
		return entry.Time.Equal(m.expected) || entry.Time.After(m.expected)
	case "lt":
		return entry.Time.Before(m.expected)
	case "lte":
		return entry.Time.Equal(m.expected) || entry.Time.Before(m.expected)
	default:
		return false
	}
}

// String returns the matcher expression.
func (m timeMatcher) String() string {
	return m.raw
}

// parseRelativeExpression parses shorthand time expressions.
func parseRelativeExpression(expr string) (Matcher, bool, error) {
	parts := strings.Fields(expr)
	if len(parts) != 2 {
		return nil, false, nil
	}

	switch strings.ToLower(parts[0]) {
	case "after":
		matcher, err := newTimeMatcher("time", "gte", parts[1], expr)
		return matcher, true, err
	case "before":
		matcher, err := newTimeMatcher("time", "lte", parts[1], expr)
		return matcher, true, err
	default:
		return nil, false, nil
	}
}

// isTimeField reports whether field should use timestamp comparison.
func isTimeField(field string) bool {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "time", "timestamp", "ts":
		return true
	default:
		return false
	}
}

// normalizeTimeOp normalizes time-specific operators.
func normalizeTimeOp(op string) string {
	switch normalizeOp(op) {
	case "after":
		return "gte"
	case "before":
		return "lte"
	default:
		return normalizeOp(op)
	}
}

// parseFilterTime parses supported filter timestamp formats.
func parseFilterTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"02/Jan/2006:15:04:05 -0700",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000",
		"15:04:05",
		"15:04",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

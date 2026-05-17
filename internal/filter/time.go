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
	timeOnly bool
	raw      string
}

// newTimeMatcher creates a timestamp matcher.
func newTimeMatcher(field string, op string, value string, raw string) (Matcher, error) {
	if !isTimeField(field) && field != "" {
		return nil, fmt.Errorf("time operators only support timestamp fields")
	}

	op = normalizeTimeOp(op)

	parsed, timeOnly, ok := parseFilterTime(value)
	if !ok {
		return nil, fmt.Errorf("invalid time filter %q", raw)
	}

	return timeMatcher{
		op:       op,
		expected: parsed,
		timeOnly: timeOnly,
		raw:      raw,
	}, nil
}

// Match reports whether the entry timestamp matches.
func (m timeMatcher) Match(entry logentry.Entry) bool {
	if !entry.HasTime {
		return false
	}

	actual := entry.Time
	expected := m.expected

	if m.timeOnly {
		actual = clockTime(entry.Time)
		expected = clockTime(m.expected)
	}

	switch m.op {
	case "gt":
		return actual.After(expected)
	case "gte":
		return actual.Equal(expected) || actual.After(expected)
	case "lt":
		return actual.Before(expected)
	case "lte":
		return actual.Equal(expected) || actual.Before(expected)
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
func parseFilterTime(value string) (time.Time, bool, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false, false
	}

	layouts := []struct {
		value    string
		timeOnly bool
	}{
		{value: time.RFC3339Nano},
		{value: time.RFC3339},
		{value: "02/Jan/2006:15:04:05 -0700"},
		{value: "2006-01-02 15:04:05"},
		{value: "2006-01-02 15:04:05.000"},
		{value: "2006-01-02T15:04:05"},
		{value: "2006-01-02T15:04:05.000"},
		{value: "15:04:05", timeOnly: true},
		{value: "15:04", timeOnly: true},
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout.value, value)
		if err == nil {
			return parsed, layout.timeOnly, true
		}
	}

	return time.Time{}, false, false
}

func clockTime(value time.Time) time.Time {
	hour, minute, second := value.Clock()

	return time.Date(0, time.January, 1, hour, minute, second, value.Nanosecond(), time.UTC)
}

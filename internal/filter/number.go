package filter

import (
	"fmt"
	"strconv"

	"github.com/gi8lino/loggo/internal/logentry"
)

// numberMatcher compares numeric field values.
type numberMatcher struct {
	field    string
	op       string
	expected float64
	raw      string
}

// newNumberMatcher creates a numeric matcher.
func newNumberMatcher(field string, op string, value string, raw string) (Matcher, error) {
	expected, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid numeric filter %q: %w", raw, err)
	}

	return numberMatcher{
		field:    field,
		op:       op,
		expected: expected,
		raw:      raw,
	}, nil
}

// Match reports whether the field number matches.
func (m numberMatcher) Match(entry logentry.Entry) bool {
	value, ok := entry.Get(m.field)
	if !ok {
		return false
	}

	actual, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}

	switch m.op {
	case "gt":
		return actual > m.expected
	case "gte":
		return actual >= m.expected
	case "lt":
		return actual < m.expected
	case "lte":
		return actual <= m.expected
	default:
		return false
	}
}

// String returns the matcher expression.
func (m numberMatcher) String() string {
	return m.raw
}

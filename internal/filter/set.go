package filter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// Set applies include and exclude matchers to log entries.
type Set struct {
	include []Matcher
	exclude []Matcher
}

// NewSet creates a filter set from profile and runtime rules.
func NewSet(p profile.Profile, includes []string, excludes []string) (*Set, error) {
	set := &Set{}

	for _, rule := range p.Filters.Include {
		matcher, err := profileRuleMatcher(rule)
		if err != nil {
			return nil, err
		}

		set.include = append(set.include, matcher)
	}

	for _, rule := range p.Filters.Exclude {
		matcher, err := profileRuleMatcher(rule)
		if err != nil {
			return nil, err
		}

		set.exclude = append(set.exclude, matcher)
	}

	for _, expr := range includes {
		matcher, err := ParseExpression(expr)
		if err != nil {
			return nil, err
		}

		set.include = append(set.include, matcher)
	}

	for _, expr := range excludes {
		matcher, err := ParseExpression(expr)
		if err != nil {
			return nil, err
		}

		set.exclude = append(set.exclude, matcher)
	}

	return set, nil
}

// Match reports whether entry should remain visible.
func (s *Set) Match(entry logentry.Entry) bool {
	for _, matcher := range s.include {
		if !matcher.Match(entry) {
			return false
		}
	}

	for _, matcher := range s.exclude {
		if matcher.Match(entry) {
			return false
		}
	}

	return true
}

// profileRuleMatcher converts a profile rule into a matcher.
func profileRuleMatcher(rule profile.Rule) (Matcher, error) {
	op := strings.TrimSpace(rule.Op)
	if op == "" {
		op = "contains"
	}

	value := profileRuleValue(rule.Value)
	raw := strings.TrimSpace(fmt.Sprintf("%s %s %s", rule.Field, op, value))

	return newMatcher(rule.Field, op, value, raw)
}

func profileRuleValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []string:
		return strings.Join(typed, ",")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, profileRuleValue(item))
		}

		return strings.Join(parts, ",")
	case fmt.Stringer:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(typed)
	default:
		return fmt.Sprint(typed)
	}
}

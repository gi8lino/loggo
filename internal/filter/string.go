package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gi8lino/loggo/internal/logentry"
)

// containsMatcher matches when a field contains a value.
type containsMatcher struct {
	field string
	value string
	raw   string
}

// Match reports whether the entry matches.
func (m containsMatcher) Match(entry logentry.Entry) bool {
	value, ok := entry.Get(m.field)

	return ok && strings.Contains(strings.ToLower(value), strings.ToLower(m.value))
}

// String returns the matcher expression.
func (m containsMatcher) String() string {
	return m.raw
}

// equalsMatcher matches when a field equals a value.
type equalsMatcher struct {
	field string
	value string
	raw   string
}

// Match reports whether the entry matches.
func (m equalsMatcher) Match(entry logentry.Entry) bool {
	value, ok := entry.Get(m.field)

	return ok && strings.EqualFold(value, m.value)
}

// String returns the matcher expression.
func (m equalsMatcher) String() string {
	return m.raw
}

// notMatcher inverts another matcher.
type notMatcher struct {
	matcher Matcher
	raw     string
}

// Match reports whether the entry does not match.
func (m notMatcher) Match(entry logentry.Entry) bool {
	return !m.matcher.Match(entry)
}

// String returns the matcher expression.
func (m notMatcher) String() string {
	return m.raw
}

// allMatcher matches when every nested matcher matches.
type allMatcher struct {
	matchers []Matcher
	raw      string
}

// Match reports whether all nested matchers match.
func (m allMatcher) Match(entry logentry.Entry) bool {
	for _, matcher := range m.matchers {
		if !matcher.Match(entry) {
			return false
		}
	}

	return true
}

// String returns the matcher expression.
func (m allMatcher) String() string {
	return m.raw
}

// anyMatcher matches when any nested matcher matches.
type anyMatcher struct {
	matchers []Matcher
	raw      string
}

// Match reports whether any nested matcher matches.
func (m anyMatcher) Match(entry logentry.Entry) bool {
	for _, matcher := range m.matchers {
		if matcher.Match(entry) {
			return true
		}
	}

	return false
}

// String returns the matcher expression.
func (m anyMatcher) String() string {
	return m.raw
}

// existsMatcher matches when a field exists.
type existsMatcher struct {
	field string
	raw   string
}

// Match reports whether the field exists.
func (m existsMatcher) Match(entry logentry.Entry) bool {
	_, ok := entry.Get(m.field)

	return ok
}

// String returns the matcher expression.
func (m existsMatcher) String() string {
	return m.raw
}

// inMatcher matches when a field is one of several values.
type inMatcher struct {
	field  string
	values []string
	raw    string
}

// Match reports whether the field value is in the configured list.
func (m inMatcher) Match(entry logentry.Entry) bool {
	value, ok := entry.Get(m.field)
	if !ok {
		return false
	}

	for _, candidate := range m.values {
		if strings.EqualFold(candidate, value) {
			return true
		}
	}

	return false
}

// String returns the matcher expression.
func (m inMatcher) String() string {
	return m.raw
}

// wildcardMatcher matches using shell-like wildcards.
type wildcardMatcher struct {
	field string
	regex *regexp.Regexp
	raw   string
}

// newWildcardMatcher creates a wildcard matcher.
func newWildcardMatcher(field string, pattern string, raw string) (Matcher, error) {
	compiled, err := regexp.Compile(wildcardPattern(pattern))
	if err != nil {
		return nil, fmt.Errorf("invalid wildcard filter %q: %w", raw, err)
	}

	return wildcardMatcher{field: field, regex: compiled, raw: raw}, nil
}

// Match reports whether the field matches the wildcard.
func (m wildcardMatcher) Match(entry logentry.Entry) bool {
	value, ok := entry.Get(m.field)

	return ok && m.regex.MatchString(value)
}

// String returns the matcher expression.
func (m wildcardMatcher) String() string {
	return m.raw
}

// regexMatcher matches using a regular expression.
type regexMatcher struct {
	field string
	regex *regexp.Regexp
	raw   string
}

// newRegexMatcher creates a regex matcher.
func newRegexMatcher(field string, pattern string, raw string) (Matcher, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex filter %q: %w", raw, err)
	}

	return regexMatcher{field: field, regex: compiled, raw: raw}, nil
}

// Match reports whether the field matches the regex.
func (m regexMatcher) Match(entry logentry.Entry) bool {
	value, ok := entry.Get(m.field)

	return ok && m.regex.MatchString(value)
}

// String returns the matcher expression.
func (m regexMatcher) String() string {
	return m.raw
}

// wildcardPattern converts a wildcard expression into a case-insensitive regexp.
func wildcardPattern(pattern string) string {
	quoted := regexp.QuoteMeta(pattern)
	quoted = strings.ReplaceAll(quoted, `\*`, `.*`)
	quoted = strings.ReplaceAll(quoted, `\?`, `.`)

	return `(?i)^` + quoted + `$`
}

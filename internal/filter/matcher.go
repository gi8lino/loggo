package filter

import "github.com/gi8lino/loggo/internal/logentry"

// Matcher decides whether a log entry matches one filter condition.
type Matcher interface {
	Match(entry logentry.Entry) bool
	String() string
}

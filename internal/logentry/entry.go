package logentry

import (
	"strings"
	"time"
)

// Entry is the normalized representation of one log line.
type Entry struct {
	Raw       string
	Parser    string
	Parsed    bool
	Timestamp string
	Time      time.Time
	HasTime   bool
	Level     string
	Message   string
	Fields    map[string]string
}

// New creates a raw log entry.
func New(raw string) Entry {
	return Entry{
		Raw:    raw,
		Fields: map[string]string{},
	}
}

// Get returns a field by name.
func (e Entry) Get(name string) (string, bool) {
	switch strings.ToLower(name) {
	case "raw":
		return e.Raw, true
	case "timestamp", "time", "ts":
		return e.Timestamp, e.Timestamp != ""
	case "level", "severity":
		return e.Level, e.Level != ""
	case "message", "msg":
		return e.Message, e.Message != ""
	}

	if value, ok := e.Fields[name]; ok {
		return value, true
	}

	lowerName := strings.ToLower(name)
	for key, value := range e.Fields {
		if strings.ToLower(key) == lowerName {
			return value, true
		}
	}

	return "", false
}

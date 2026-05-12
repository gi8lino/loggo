package parser

import (
	"strconv"
	"strings"
	"time"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
)

// applyCommonFields extracts timestamp, level, and message fields.
func applyCommonFields(entry *logentry.Entry, p profile.Profile) {
	entry.Timestamp = firstField(entry.Fields, p.TimestampField, "time", "timestamp", "ts", "@timestamp")
	entry.Level = normalizeLevel(firstField(entry.Fields, p.LevelField, "level", "severity", "lvl"))
	entry.Message = firstField(entry.Fields, p.MessageField, "msg", "message", "log")

	if parsed, ok := ParseTime(entry.Timestamp); ok {
		entry.Time = parsed
		entry.HasTime = true
	}
}

// ParseTime parses common log timestamp formats.
func ParseTime(value string) (time.Time, bool) {
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
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, true
		}
	}

	if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
		switch {
		case unix > 1_000_000_000_000:
			return time.UnixMilli(unix), true
		case unix > 1_000_000_000:
			return time.Unix(unix, 0), true
		}
	}

	return time.Time{}, false
}

// firstField returns the first existing field value.
func firstField(fields map[string]string, names ...string) string {
	for _, name := range names {
		if name == "" {
			continue
		}

		if value, ok := fields[name]; ok && value != "" {
			return value
		}

		lowerName := strings.ToLower(name)
		for key, value := range fields {
			if strings.ToLower(key) == lowerName && value != "" {
				return value
			}
		}
	}

	return ""
}

// normalizeLevel normalizes common severity names.
func normalizeLevel(level string) string {
	level = strings.TrimSpace(strings.ToUpper(level))

	switch level {
	case "WARNING":
		return "WARN"
	case "ERR":
		return "ERROR"
	default:
		return level
	}
}

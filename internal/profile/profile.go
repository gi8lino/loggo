package profile

import (
	"maps"
	"slices"
	"strings"
)

// Parser names.
const (
	ParserAuto   = "auto"
	ParserJSON   = "json"
	ParserLogfmt = "logfmt"
	ParserRegex  = "regex"
	ParserSplit  = "split"
	ParserRaw    = "raw"
)

// Profile describes how a log stream should be parsed, filtered, and rendered.
type Profile struct {
	Name           string      `yaml:"-"`
	Parser         string      `yaml:"parser"`
	Regex          string      `yaml:"regex"`
	TimestampField string      `yaml:"timestampField"`
	LevelField     string      `yaml:"levelField"`
	MessageField   string      `yaml:"messageField"`
	Fields         []string    `yaml:"fields"`
	FixedFields    bool        `yaml:"fixedFields"`
	HiddenFields   []string    `yaml:"hiddenFields"`
	Format         string      `yaml:"format"`
	Split          SplitConfig `yaml:"split"`
	Filters        Filters     `yaml:"filters"`
	Colors         Colors      `yaml:"colors"`
}

// SplitConfig describes positional log parsing.
type SplitConfig struct {
	Delimiter string   `yaml:"delimiter"`
	Fields    []string `yaml:"fields"`
}

// Filters describes profile-level include and exclude filters.
type Filters struct {
	Include []Rule `yaml:"include"`
	Exclude []Rule `yaml:"exclude"`
}

// Rule describes a field-aware filter rule.
type Rule struct {
	Field string `yaml:"field"`
	Op    string `yaml:"op"`
	Value any    `yaml:"value"`
	Expr  string `yaml:"expr,omitempty"`
}

// Colors describes color output rules.
type Colors struct {
	Levels    map[string]string `yaml:"levels"`
	Fields    map[string]string `yaml:"fields"`
	Timestamp string            `yaml:"timestamp"`
	Message   string            `yaml:"message"`
}

// RuntimeOverrides contains CLI-level profile overrides.
type RuntimeOverrides struct {
	Parser       string
	Split        string
	Fields       []string
	HiddenFields []string
	Format       string
}

// Builtins returns built-in profiles.
func Builtins() map[string]Profile {
	return map[string]Profile{
		"auto": Normalize("auto", Profile{
			Parser: ParserAuto,
			Fields: []string{
				"service",
				"component",
				"request_id",
				"trace_id",
				"method",
				"path",
				"status",
				"duration",
			},
		}),
		"json": Normalize("json", Profile{
			Parser: ParserJSON,
			Fields: []string{
				"service",
				"request_id",
				"trace_id",
				"method",
				"path",
				"status",
				"duration",
			},
		}),
		"logfmt": Normalize("logfmt", Profile{
			Parser: ParserLogfmt,
			Fields: []string{
				"service",
				"component",
				"method",
				"path",
				"status",
				"duration",
			},
		}),
		"zap": Normalize("zap", Profile{
			Parser:         ParserJSON,
			TimestampField: "ts",
			LevelField:     "level",
			MessageField:   "msg",
			Fields: []string{
				"logger",
				"caller",
				"service",
				"trace_id",
				"span_id",
				"method",
				"path",
				"status",
				"duration",
			},
		}),
		"ecs": Normalize("ecs", Profile{
			Parser:         ParserJSON,
			TimestampField: "@timestamp",
			LevelField:     "log.level",
			MessageField:   "message",
			Fields: []string{
				"service.name",
				"event.dataset",
				"trace.id",
				"transaction.id",
				"http.request.method",
				"url.path",
				"http.response.status_code",
			},
		}),
		"cri": Normalize("cri", Profile{
			Parser:         ParserSplit,
			TimestampField: "time",
			MessageField:   "message",
			Split: SplitConfig{
				Delimiter: " ",
				Fields:    []string{"time", "stream", "flags", "message"},
			},
			Fields: []string{
				"stream",
				"flags",
			},
		}),
		"nginx": Normalize("nginx", Profile{
			Parser:         ParserRegex,
			Regex:          `^(?P<remote_addr>\S+) (?P<remote_ident>\S+) (?P<remote_user>\S+) \[(?P<time>[^\]]+)\] "(?P<method>\S+) (?P<path>\S+)(?: (?P<protocol>[^"]+))?" (?P<status>\d+) (?P<bytes>\d+) "(?P<referer>[^"]*)" "(?P<user_agent>[^"]*)"(?: "(?P<forwarded_for>[^"]*)")?`,
			TimestampField: "time",
			MessageField:   "path",
			Fields: []string{
				"remote_addr",
				"remote_user",
				"method",
				"path",
				"status",
				"bytes",
				"user_agent",
				"forwarded_for",
			},
			Filters: Filters{
				Exclude: []Rule{
					{Field: "path", Op: "equals", Value: "/status.php"},
					{Field: "user_agent", Op: "wildcard", Value: "*kube-probe*"},
				},
			},
		}),
		"apache": Normalize("apache", Profile{
			Parser:         ParserRegex,
			Regex:          `^(?P<remote_addr>\S+) (?P<remote_ident>\S+) (?P<remote_user>\S+) \[(?P<time>[^\]]+)\] "(?P<method>\S+) (?P<path>\S+)(?: (?P<protocol>[^"]+))?" (?P<status>\d+) (?P<bytes>\d+|-) "(?P<referer>[^"]*)" "(?P<user_agent>[^"]*)"`,
			TimestampField: "time",
			MessageField:   "path",
			Fields: []string{
				"remote_addr",
				"remote_user",
				"method",
				"path",
				"status",
				"bytes",
				"user_agent",
			},
		}),
		"postgres": Normalize("postgres", Profile{
			Parser:         ParserRegex,
			Regex:          `^(?P<time>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?) (?P<timezone>[A-Z]{2,5}) \[(?P<pid>\d+)\] (?:(?P<db_user>[^@ ]+)@(?P<database>[^ ]+) )?(?P<level>[A-Z]+): (?P<message>.*)$`,
			TimestampField: "time",
			LevelField:     "level",
			MessageField:   "message",
			Fields: []string{
				"pid",
				"db_user",
				"database",
			},
		}),
		"raw": Normalize("raw", Profile{
			Parser: ParserRaw,
		}),
	}
}

// Normalize applies safe defaults to a profile.
func Normalize(name string, p Profile) Profile {
	p.Name = name

	if p.Parser == "" {
		p.Parser = ParserAuto
	}
	if p.TimestampField == "" {
		p.TimestampField = "time"
	}
	if p.LevelField == "" {
		p.LevelField = "level"
	}
	if p.MessageField == "" {
		p.MessageField = "msg"
	}
	if p.Split.Delimiter == "" {
		p.Split.Delimiter = " "
	}
	if p.Colors.Timestamp == "" {
		p.Colors.Timestamp = "dim"
	}
	if p.Colors.Message == "" {
		p.Colors.Message = "reset"
	}

	p.HiddenFields = normalizeFields(p.HiddenFields)
	p.Colors.Levels = mergeLevelColors(p.Colors.Levels)

	if p.Colors.Fields == nil {
		p.Colors.Fields = map[string]string{}
	}

	return p
}

// WithRuntimeOverrides applies CLI-level overrides to a profile.
func (p Profile) WithRuntimeOverrides(overrides RuntimeOverrides) Profile {
	if overrides.Parser != "" {
		p.Parser = overrides.Parser
	}
	if overrides.Split != "" {
		p.Parser = ParserSplit
		p.Split.Delimiter = overrides.Split
	}
	if len(overrides.Fields) > 0 {
		p.Fields = overrides.Fields
	}
	if len(overrides.HiddenFields) > 0 {
		p.HiddenFields = mergeFields(p.HiddenFields, overrides.HiddenFields)
	}
	if overrides.Format != "" {
		p.Format = overrides.Format
	}

	return Normalize(p.Name, p)
}

// mergeLevelColors merges configured level colors with defaults.
func mergeLevelColors(configured map[string]string) map[string]string {
	levels := map[string]string{
		"TRACE": "dim",
		"DEBUG": "dim",
		"INFO":  "cyan",
		"WARN":  "yellow",
		"ERROR": "red",
		"FATAL": "magenta",
		"PANIC": "magenta",
	}

	normalized := map[string]string{}
	for level, color := range configured {
		normalized[strings.ToUpper(level)] = color
	}

	maps.Copy(levels, normalized)

	return levels
}

// normalizeFields trims, deduplicates, and preserves field order.
func normalizeFields(fields []string) []string {
	seen := map[string]struct{}{}
	normalized := []string{}

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}

		seen[field] = struct{}{}
		normalized = append(normalized, field)
	}

	return normalized
}

// mergeFields merges two field lists and preserves first-seen order.
func mergeFields(base []string, extra []string) []string {
	return normalizeFields(slices.Concat(base, extra))
}

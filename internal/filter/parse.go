package filter

import (
	"fmt"
	"strings"
)

// ParseExpression parses a human-friendly filter expression.
func ParseExpression(expr string) (Matcher, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("filter expression must not be empty")
	}

	if matcher, ok, err := parseRelativeExpression(expr); ok || err != nil {
		return matcher, err
	}

	for _, op := range []string{">=", "<=", "!=", "=", ">", "<"} {
		if idx := strings.Index(expr, op); idx > 0 {
			field := strings.TrimSpace(expr[:idx])
			value := strings.TrimSpace(expr[idx+len(op):])

			if field == "" || value == "" {
				return nil, fmt.Errorf("invalid filter expression %q", expr)
			}

			return newMatcher(field, normalizeOp(op), value, expr)
		}
	}

	parts := strings.Fields(expr)
	if len(parts) >= 3 {
		op := normalizeOp(parts[1])
		if isWordOp(op) {
			return newMatcher(parts[0], op, strings.Join(parts[2:], " "), expr)
		}
	}

	if len(parts) == 2 && normalizeOp(parts[1]) == "exists" {
		return newMatcher(parts[0], "exists", "", expr)
	}

	return newMatcher("raw", "contains", expr, expr)
}

// newMatcher creates a matcher from parsed expression parts.
func newMatcher(field string, op string, value string, raw string) (Matcher, error) {
	field = strings.TrimSpace(field)
	op = normalizeOp(op)
	value = strings.TrimSpace(value)

	if isTimeField(field) || op == "after" || op == "before" {
		return newTimeMatcher(field, op, value, raw)
	}

	switch op {
	case "", "contains":
		return containsMatcher{field: field, value: value, raw: raw}, nil
	case "wildcard":
		return newWildcardMatcher(field, value, raw)
	case "regex":
		return newRegexMatcher(field, value, raw)
	case "eq":
		return equalsMatcher{field: field, value: value, raw: raw}, nil
	case "neq":
		return notMatcher{matcher: equalsMatcher{field: field, value: value, raw: raw}, raw: raw}, nil
	case "gt", "gte", "lt", "lte":
		return newNumberMatcher(field, op, value, raw)
	case "exists":
		return existsMatcher{field: field, raw: raw}, nil
	case "in":
		return inMatcher{field: field, values: splitValues(value), raw: raw}, nil
	default:
		return nil, fmt.Errorf("unknown filter operator %q", op)
	}
}

// normalizeOp normalizes symbolic and word operators.
func normalizeOp(op string) string {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "=":
		return "eq"
	case "==":
		return "eq"
	case "equals":
		return "eq"
	case "eq":
		return "eq"
	case "!=":
		return "neq"
	case "not_equals":
		return "neq"
	case "neq":
		return "neq"
	case ">":
		return "gt"
	case ">=":
		return "gte"
	case "<":
		return "lt"
	case "<=":
		return "lte"
	case "contains":
		return "contains"
	case "wildcard", "glob":
		return "wildcard"
	case "regex":
		return "regex"
	case "exists":
		return "exists"
	case "in":
		return "in"
	case "after":
		return "after"
	case "before":
		return "before"
	default:
		return strings.ToLower(strings.TrimSpace(op))
	}
}

// isWordOp reports whether op can appear between a field and value.
func isWordOp(op string) bool {
	switch op {
	case "contains", "wildcard", "regex", "eq", "neq", "gt", "gte", "lt", "lte", "in", "after", "before":
		return true
	default:
		return false
	}
}

// splitValues splits a comma-separated value list.
func splitValues(value string) []string {
	values := []string{}

	for item := range strings.SplitSeq(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		values = append(values, item)
	}

	return values
}

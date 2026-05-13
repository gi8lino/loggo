package filter

import (
	"testing"

	"github.com/gi8lino/loggo/internal/logentry"
)

func TestParseExpressionSupportsBooleanGroups(t *testing.T) {
	matcher, err := ParseExpression(`level = ERROR and (status >= 500 or path wildcard /admin/*)`)
	if err != nil {
		t.Fatalf("ParseExpression returned error: %v", err)
	}

	matching := logentry.New(`{"level":"ERROR","status":"503","path":"/health"}`)
	matching.Level = "ERROR"
	matching.Fields["status"] = "503"
	matching.Fields["path"] = "/health"

	if !matcher.Match(matching) {
		t.Fatalf("expected matcher to match grouped expression")
	}

	nonMatching := logentry.New(`{"level":"INFO","status":"503","path":"/admin/users"}`)
	nonMatching.Level = "INFO"
	nonMatching.Fields["status"] = "503"
	nonMatching.Fields["path"] = "/admin/users"

	if matcher.Match(nonMatching) {
		t.Fatalf("expected matcher to reject entry outside grouped expression")
	}
}

func TestParseExpressionSupportsNegation(t *testing.T) {
	matcher, err := ParseExpression(`not (path wildcard /health* or path wildcard /metrics*)`)
	if err != nil {
		t.Fatalf("ParseExpression returned error: %v", err)
	}

	visible := logentry.New("request")
	visible.Fields["path"] = "/orders/42"
	if !matcher.Match(visible) {
		t.Fatalf("expected negated matcher to keep non-excluded path")
	}

	hidden := logentry.New("request")
	hidden.Fields["path"] = "/metrics/prometheus"
	if matcher.Match(hidden) {
		t.Fatalf("expected negated matcher to reject excluded path")
	}
}

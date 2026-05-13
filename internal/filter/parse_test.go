package filter

import (
	"testing"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseExpressionSupportsBooleanGroups(t *testing.T) {
	matcher, err := ParseExpression(`level = ERROR and (status >= 500 or path wildcard /admin/*)`)
	require.NoError(t, err)

	matching := logentry.New(`{"level":"ERROR","status":"503","path":"/health"}`)
	matching.Level = "ERROR"
	matching.Fields["status"] = "503"
	matching.Fields["path"] = "/health"

	assert.True(t, matcher.Match(matching))

	nonMatching := logentry.New(`{"level":"INFO","status":"503","path":"/admin/users"}`)
	nonMatching.Level = "INFO"
	nonMatching.Fields["status"] = "503"
	nonMatching.Fields["path"] = "/admin/users"

	assert.False(t, matcher.Match(nonMatching))
}

func TestParseExpressionSupportsNegation(t *testing.T) {
	matcher, err := ParseExpression(`not (path wildcard /health* or path wildcard /metrics*)`)
	require.NoError(t, err)

	visible := logentry.New("request")
	visible.Fields["path"] = "/orders/42"
	assert.True(t, matcher.Match(visible))

	hidden := logentry.New("request")
	hidden.Fields["path"] = "/metrics/prometheus"
	assert.False(t, matcher.Match(hidden))
}

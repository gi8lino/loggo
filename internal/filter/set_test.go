package filter

import (
	"testing"

	"github.com/gi8lino/loggo/internal/logentry"
	"github.com/gi8lino/loggo/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSetSupportsProfileRulesWithSpaces(t *testing.T) {
	set, err := NewSet(profile.Normalize("spaces", profile.Profile{
		Filters: profile.Filters{
			Include: []profile.Rule{
				{Field: "message", Op: "equals", Value: "health check passed"},
			},
		},
	}), nil, nil)
	require.NoError(t, err)

	matching := logentry.New("matching")
	matching.Message = "health check passed"
	assert.True(t, set.Match(matching))

	nonMatching := logentry.New("non-matching")
	nonMatching.Message = "health check failed"
	assert.False(t, set.Match(nonMatching))
}

func TestNewSetSupportsProfileRegexRulesWithSpaces(t *testing.T) {
	set, err := NewSet(profile.Normalize("regex", profile.Profile{
		Filters: profile.Filters{
			Include: []profile.Rule{
				{Field: "message", Op: "regex", Value: `^GET /admin/users \(\d+\)$`},
			},
		},
	}), nil, nil)
	require.NoError(t, err)

	matching := logentry.New("matching")
	matching.Message = "GET /admin/users (42)"
	assert.True(t, set.Match(matching))

	nonMatching := logentry.New("non-matching")
	nonMatching.Message = "GET /admin/users"
	assert.False(t, set.Match(nonMatching))
}

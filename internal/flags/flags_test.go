package flags

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestParseFlagsOverrideValues preserves plain diagnostic values and CLI precedence.
func TestParseFlagsOverrideValues(t *testing.T) {
	t.Setenv("LOGGO_PROFILE", "environment")
	options, err := ParseFlags([]string{"--profile", "command-line"}, "test")
	require.NoError(t, err)
	assert.Equal(t, "command-line", options.Profile)
	assert.Equal(t, "command-line", options.OverriddenValues["profile"])
}

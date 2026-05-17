package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMergesBuiltinProfileOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "loggo.yaml")

	content := []byte(`
profiles:
  nginx:
    fields:
      - status
    hiddenFields:
      - user_agent
`)
	require.NoError(t, os.WriteFile(path, content, 0o600))

	cfg, _, err := Load([]string{path}, func(string) string { return "" })
	require.NoError(t, err)

	nginx, err := cfg.ResolveProfile("nginx", func(string) string { return "" })
	require.NoError(t, err)

	assert.Equal(t, "regex", nginx.Parser)
	assert.NotEmpty(t, nginx.Regex)
	assert.Equal(t, []string{"status"}, nginx.Fields)
	assert.Equal(t, []string{"user_agent"}, nginx.HiddenFields)
}

func TestLoadMergesCustomProfilesAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "base.yaml")
	overridePath := filepath.Join(dir, "override.yaml")

	baseContent := []byte(`
profiles:
  app:
    parser: json
    timestampField: ts
    levelField: severity
`)
	overrideContent := []byte(`
profiles:
  app:
    hiddenFields:
      - secret
`)

	require.NoError(t, os.WriteFile(basePath, baseContent, 0o600))
	require.NoError(t, os.WriteFile(overridePath, overrideContent, 0o600))

	cfg, _, err := Load([]string{basePath, overridePath}, func(string) string { return "" })
	require.NoError(t, err)

	app, err := cfg.ResolveProfile("app", func(string) string { return "" })
	require.NoError(t, err)

	assert.Equal(t, "json", app.Parser)
	assert.Equal(t, "ts", app.TimestampField)
	assert.Equal(t, "severity", app.LevelField)
	assert.Equal(t, []string{"secret"}, app.HiddenFields)
}

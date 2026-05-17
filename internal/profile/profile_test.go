package profile_test

import (
	"testing"

	logparser "github.com/gi8lino/loggo/internal/parser"
	"github.com/gi8lino/loggo/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinsIncludeAdditionalCommonProfiles(t *testing.T) {
	builtins := profile.Builtins()

	for _, name := range []string{"apache", "postgres", "zap", "ecs", "cri"} {
		_, ok := builtins[name]
		assert.True(t, ok, "expected built-in profile %q", name)
	}
}

func TestBuiltinApacheProfileParsesCombinedLogs(t *testing.T) {
	parser, err := logparser.New(profile.Builtins()["apache"])
	require.NoError(t, err)

	entry := parser.Parse(`203.0.113.10 - frank [12/May/2026:13:14:31 +0000] "GET /index.html HTTP/1.1" 200 1234 "https://example.com" "curl/8.7.1"`)

	assert.True(t, entry.Parsed)
	assert.Equal(t, "GET", entry.Fields["method"])
	assert.Equal(t, "/index.html", entry.Message)
	assert.Equal(t, "200", entry.Fields["status"])
}

func TestBuiltinPostgresProfileParsesDefaultLogs(t *testing.T) {
	parser, err := logparser.New(profile.Builtins()["postgres"])
	require.NoError(t, err)

	entry := parser.Parse(`2026-05-17 13:14:31.123 UTC [4242] app@billing ERROR: duplicate key value violates unique constraint`)

	assert.True(t, entry.Parsed)
	assert.Equal(t, "ERROR", entry.Level)
	assert.Equal(t, "4242", entry.Fields["pid"])
	assert.Equal(t, "app", entry.Fields["db_user"])
	assert.Equal(t, "billing", entry.Fields["database"])
}

func TestBuiltinZapProfileParsesStructuredJSON(t *testing.T) {
	parser, err := logparser.New(profile.Builtins()["zap"])
	require.NoError(t, err)

	entry := parser.Parse(`{"ts":"2026-05-17T13:14:31Z","level":"error","logger":"payments","caller":"api/handler.go:42","msg":"charge failed","status":502}`)

	assert.True(t, entry.Parsed)
	assert.Equal(t, "ERROR", entry.Level)
	assert.Equal(t, "charge failed", entry.Message)
	assert.Equal(t, "payments", entry.Fields["logger"])
}

func TestBuiltinECSProfileParsesElasticCommonSchemaJSON(t *testing.T) {
	parser, err := logparser.New(profile.Builtins()["ecs"])
	require.NoError(t, err)

	entry := parser.Parse(`{"@timestamp":"2026-05-17T13:14:31Z","log.level":"warn","message":"request slow","service.name":"checkout","http.request.method":"POST","url.path":"/orders","http.response.status_code":504}`)

	assert.True(t, entry.Parsed)
	assert.Equal(t, "WARN", entry.Level)
	assert.Equal(t, "request slow", entry.Message)
	assert.Equal(t, "checkout", entry.Fields["service.name"])
}

func TestBuiltinCRIProfileParsesContainerRuntimeLogs(t *testing.T) {
	parser, err := logparser.New(profile.Builtins()["cri"])
	require.NoError(t, err)

	entry := parser.Parse(`2026-05-17T13:14:31.123456789Z stdout F worker finished job=42`)

	assert.True(t, entry.Parsed)
	assert.Equal(t, "stdout", entry.Fields["stream"])
	assert.Equal(t, "F", entry.Fields["flags"])
	assert.Equal(t, "worker finished job=42", entry.Message)
}

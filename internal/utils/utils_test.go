package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CreateUsers", "create_users"},
		{"getHTTPResponse", "get_http_response"},
		{"simpleTest", "simple_test"},
		{"already_snake", "already_snake"},
		{"A", "a"},
		{"", ""},
		{"JSONParser", "json_parser"},
		{"myURL", "my_url"},
		{"ID", "id"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, ToSnakeCase(tc.input))
		})
	}
}

func TestTimestampPrefix(t *testing.T) {
	ts := TimestampPrefix()
	// Must be exactly 14 digits matching YYYYMMDDHHMMSS.
	assert.Len(t, ts, 14)
	assert.Regexp(t, regexp.MustCompile(`^\d{14}$`), ts)
}

func TestFileExists(t *testing.T) {
	// Existing file
	tmp := t.TempDir()
	f := filepath.Join(tmp, "exists.txt")
	require.NoError(t, os.WriteFile(f, []byte("hi"), 0o644))
	assert.True(t, FileExists(f))

	// Non-existing file
	assert.False(t, FileExists(filepath.Join(tmp, "nope.txt")))

	// Directory counts as existing
	assert.True(t, FileExists(tmp))
}

func TestEnsureDir(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "a", "b", "c")

	require.NoError(t, EnsureDir(nested))
	assert.True(t, FileExists(nested))

	// Calling again is idempotent
	require.NoError(t, EnsureDir(nested))
}

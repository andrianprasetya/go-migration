package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterpolateEnv_BasicReplacement(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")

	input := []byte("host: ${DB_HOST}\nport: ${DB_PORT}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "host: localhost\nport: 5432", string(result))
}

func TestInterpolateEnv_PartialStringReplacement(t *testing.T) {
	t.Setenv("REGION", "us-east-1")

	input := []byte(`host: "db-${REGION}.example.com"`)
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, `host: "db-us-east-1.example.com"`, string(result))
}

func TestInterpolateEnv_UnresolvedVariable(t *testing.T) {
	// Ensure the variable is not set
	os.Unsetenv("MISSING_VAR")

	input := []byte("password: ${MISSING_VAR}")
	_, err := InterpolateEnv(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MISSING_VAR")
	assert.Contains(t, err.Error(), "unresolved environment variables")
}

func TestInterpolateEnv_MultipleUnresolved(t *testing.T) {
	os.Unsetenv("VAR_A")
	os.Unsetenv("VAR_B")

	input := []byte("a: ${VAR_A}\nb: ${VAR_B}")
	_, err := InterpolateEnv(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "VAR_A")
	assert.Contains(t, err.Error(), "VAR_B")
}

func TestInterpolateEnv_EscapedPlaceholder(t *testing.T) {
	// $${VAR} should become literal ${VAR}, not interpolated
	t.Setenv("SECRET", "should_not_appear")

	input := []byte("literal: $${SECRET}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "literal: ${SECRET}", string(result))
}

func TestInterpolateEnv_EscapedAndRegularMixed(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")

	input := []byte("host: ${DB_HOST}\nexample: $${DB_HOST}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "host: localhost\nexample: ${DB_HOST}", string(result))
}

func TestInterpolateEnv_NoPlaceholders(t *testing.T) {
	input := []byte("host: localhost\nport: 5432")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "host: localhost\nport: 5432", string(result))
}

func TestInterpolateEnv_EmptyInput(t *testing.T) {
	result, err := InterpolateEnv([]byte(""))
	require.NoError(t, err)
	assert.Equal(t, "", string(result))
}

func TestInterpolateEnv_EmptyValue(t *testing.T) {
	t.Setenv("EMPTY_VAR", "")

	input := []byte("val: ${EMPTY_VAR}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "val: ", string(result))
}

func TestInterpolateEnv_UnderscoreVarName(t *testing.T) {
	t.Setenv("_PRIVATE_VAR", "secret")

	input := []byte("val: ${_PRIVATE_VAR}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "val: secret", string(result))
}

func TestInterpolateEnv_MultipleEscaped(t *testing.T) {
	input := []byte("a: $${FOO}\nb: $${BAR}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "a: ${FOO}\nb: ${BAR}", string(result))
}

func TestInterpolateEnv_EscapedUnsetVarNoError(t *testing.T) {
	// Escaped placeholders should not trigger unresolved errors
	os.Unsetenv("UNSET_VAR")

	input := []byte("val: $${UNSET_VAR}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "val: ${UNSET_VAR}", string(result))
}

func TestInterpolateEnv_SameVarMultipleTimes(t *testing.T) {
	t.Setenv("REPEATED", "value")

	input := []byte("a: ${REPEATED} b: ${REPEATED}")
	result, err := InterpolateEnv(input)
	require.NoError(t, err)
	assert.Equal(t, "a: value b: value", string(result))
}

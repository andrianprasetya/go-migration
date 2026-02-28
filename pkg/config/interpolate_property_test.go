package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// genEnvVarName generates a valid environment variable name with a unique prefix
// to avoid collisions with real system env vars.
func genEnvVarName() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		// Use uppercase-only names to avoid case-insensitive collisions on Windows,
		// where env var names like "GOMIG_TEST_d" and "GOMIG_TEST_D" are identical.
		suffix := rapid.StringMatching(`[A-Z][A-Z0-9_]{0,12}`).Draw(t, "suffix")
		return "GOMIG_TEST_" + suffix
	})
}

// genEnvVarValue generates a random string value for an environment variable.
// Values avoid containing ${...} patterns to keep assertions clean.
func genEnvVarValue() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		return rapid.StringMatching(`[A-Za-z0-9 _\-\./,:;@#%^&*()]{0,50}`).Draw(t, "value")
	})
}

// genUniqueEnvVarNames generates a slice of unique env var names.
func genUniqueEnvVarNames(minN, maxN int) *rapid.Generator[[]string] {
	return rapid.Custom(func(t *rapid.T) []string {
		n := rapid.IntRange(minN, maxN).Draw(t, "varCount")
		seen := make(map[string]bool)
		names := make([]string, 0, n)
		for len(names) < n {
			name := genEnvVarName().Draw(t, "varName")
			// Use case-folded key for uniqueness check (Windows env vars are case-insensitive)
			upper := strings.ToUpper(name)
			if !seen[upper] {
				seen[upper] = true
				names = append(names, name)
			}
		}
		return names
	})
}

// setAndCleanupEnvVars sets environment variables and returns a cleanup function.
func setAndCleanupEnvVars(vars map[string]string) func() {
	for k, v := range vars {
		os.Setenv(k, v)
	}
	return func() {
		for k := range vars {
			os.Unsetenv(k)
		}
	}
}

// Feature: library-improvements, Property 13: Environment variable interpolation replacement
// **Validates: Requirements 8.1**

func TestProperty13_EnvVarInterpolationReplacement(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate unique env var names and values.
		names := genUniqueEnvVarNames(1, 5).Draw(t, "names")
		values := make(map[string]string, len(names))
		for _, name := range names {
			val := genEnvVarValue().Draw(t, "val_"+name)
			values[name] = val
		}

		// Set all env vars and defer cleanup.
		cleanup := setAndCleanupEnvVars(values)
		defer cleanup()

		// Build input with ${VAR_NAME} placeholders, optionally with surrounding text.
		var inputParts []string
		for _, name := range names {
			prefix := rapid.StringMatching(`[a-z]{0,10}`).Draw(t, "prefix_"+name)
			suffix := rapid.StringMatching(`[a-z]{0,10}`).Draw(t, "suffix_"+name)
			inputParts = append(inputParts, fmt.Sprintf("%s${%s}%s", prefix, name, suffix))
		}
		input := strings.Join(inputParts, "\n")

		result, err := InterpolateEnv([]byte(input))
		require.NoError(t, err, "InterpolateEnv should not error when all vars are set")

		output := string(result)

		// Verify each ${VAR_NAME} was replaced with its value.
		for _, name := range names {
			assert.Contains(t, output, values[name],
				"output should contain the value of %s", name)
		}

		// Verify no ${VAR_NAME} patterns remain for the set variables.
		for _, name := range names {
			placeholder := fmt.Sprintf("${%s}", name)
			assert.NotContains(t, output, placeholder,
				"output should not contain the placeholder %s for a set variable", placeholder)
		}
	})
}

// Feature: library-improvements, Property 14: Unresolved environment variable error
// **Validates: Requirements 8.2**

func TestProperty14_UnresolvedEnvVarError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate unique env var names that we will NOT set.
		names := genUniqueEnvVarNames(1, 5).Draw(t, "names")

		// Ensure none of the generated vars are set.
		for _, name := range names {
			os.Unsetenv(name)
		}

		// Build input with ${VAR_NAME} placeholders.
		var inputParts []string
		for _, name := range names {
			inputParts = append(inputParts, fmt.Sprintf("key: ${%s}", name))
		}
		input := strings.Join(inputParts, "\n")

		_, err := InterpolateEnv([]byte(input))

		// Verify a non-nil error is returned.
		require.Error(t, err, "InterpolateEnv should return an error for unset variables")

		// Verify the error message contains each unresolved variable name.
		for _, name := range names {
			assert.Contains(t, err.Error(), name,
				"error message should contain the unresolved variable name %s", name)
		}
	})
}

// Feature: library-improvements, Property 15: Escaped environment variable literal
// **Validates: Requirements 8.5**

func TestProperty15_EscapedEnvVarLiteral(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate unique env var names.
		names := genUniqueEnvVarNames(1, 5).Draw(t, "names")

		// Ensure the vars are NOT set so we can confirm no lookup is attempted.
		for _, name := range names {
			os.Unsetenv(name)
		}

		// Build input with $${VAR_NAME} escape sequences.
		var inputParts []string
		for _, name := range names {
			prefix := rapid.StringMatching(`[a-z]{0,10}`).Draw(t, "prefix_"+name)
			suffix := rapid.StringMatching(`[a-z]{0,10}`).Draw(t, "suffix_"+name)
			inputParts = append(inputParts, fmt.Sprintf("%s$${%s}%s", prefix, name, suffix))
		}
		input := strings.Join(inputParts, "\n")

		result, err := InterpolateEnv([]byte(input))

		// Escaped vars should NOT trigger an unresolved error.
		require.NoError(t, err, "InterpolateEnv should not error for escaped placeholders")

		output := string(result)

		// Verify each $${VAR_NAME} was replaced with the literal ${VAR_NAME}.
		for _, name := range names {
			literal := fmt.Sprintf("${%s}", name)
			assert.Contains(t, output, literal,
				"output should contain the literal text %s for escaped placeholder", literal)
		}

		// Verify no $${VAR_NAME} escape sequences remain in the output.
		for _, name := range names {
			escaped := fmt.Sprintf("$${%s}", name)
			assert.NotContains(t, output, escaped,
				"output should not contain the escape sequence %s", escaped)
		}
	})
}

package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// envVarPattern matches ${VAR_NAME} placeholders where VAR_NAME is a valid
// identifier (starts with letter or underscore, followed by alphanumerics or underscores).
var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// escapePattern matches $${...} escape sequences — a double dollar sign before a brace placeholder.
var escapePattern = regexp.MustCompile(`\$\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// InterpolateEnv replaces ${VAR_NAME} placeholders in data with the corresponding
// environment variable values. Escaped sequences ($${VAR_NAME}) are replaced with
// the literal text ${VAR_NAME}. Returns an error listing all unresolved variables.
func InterpolateEnv(data []byte) ([]byte, error) {
	s := string(data)

	// Step 1: Replace escape sequences $${VAR} with a temporary sentinel that
	// won't be matched by the interpolation regex.
	const sentinel = "\x00ESCAPED_ENV\x00"
	type escapedEntry struct {
		varName string
	}
	var escaped []escapedEntry

	s = escapePattern.ReplaceAllStringFunc(s, func(match string) string {
		sub := escapePattern.FindStringSubmatch(match)
		escaped = append(escaped, escapedEntry{varName: sub[1]})
		return fmt.Sprintf("%s%d%s", sentinel, len(escaped)-1, sentinel)
	})

	// Step 2: Find all ${VAR} placeholders and collect unresolved ones.
	var unresolved []string
	result := envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		sub := envVarPattern.FindStringSubmatch(match)
		varName := sub[1]
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		unresolved = append(unresolved, varName)
		return match // leave placeholder as-is for error reporting
	})

	// Step 3: Restore escaped sequences to literal ${VAR_NAME}.
	for i, entry := range escaped {
		placeholder := fmt.Sprintf("%s%d%s", sentinel, i, sentinel)
		result = strings.Replace(result, placeholder, "${"+entry.varName+"}", 1)
	}

	if len(unresolved) > 0 {
		return nil, fmt.Errorf("unresolved environment variables: %s", strings.Join(unresolved, ", "))
	}

	return []byte(result), nil
}

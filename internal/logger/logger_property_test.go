package logger

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// allLevels enumerates every LogLevel from Debug to Error.
var allLevels = []LogLevel{LevelDebug, LevelInfo, LevelWarn, LevelError}

// levelGenerator returns a rapid generator that picks one of the four log levels.
func levelGenerator() *rapid.Generator[LogLevel] {
	return rapid.SampledFrom(allLevels)
}

// directionGenerator returns "up" or "down".
func directionGenerator() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"up", "down"})
}

// resultGenerator returns "success" or "failure".
func resultGenerator() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"success", "failure"})
}

// migrationNameGenerator produces realistic migration names.
func migrationNameGenerator() *rapid.Generator[string] {
	return rapid.StringMatching(`[0-9]{14}_[a-z_]{3,20}`)
}

// Feature: go-migration, Property 37: Migration execution is logged with required fields
// **Validates: Requirements 18.1**
//
// For any migration execution (up or down, success or failure), the logger
// should receive a log entry containing the migration name, direction, and result.
func TestProperty37_MigrationExecutionLoggedWithRequiredFields(t *testing.T) {
	t.Run("ConsoleLogger", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			name := migrationNameGenerator().Draw(t, "migrationName")
			direction := directionGenerator().Draw(t, "direction")
			result := resultGenerator().Draw(t, "result")

			var buf bytes.Buffer
			lg := newConsoleLoggerWithWriter(LevelDebug, &buf)

			// Simulate a migration log entry the way the migrator would log it.
			lg.Info("migration %s %s %s", name, direction, result)

			output := buf.String()
			assert.Contains(t, output, name,
				"log output should contain migration name")
			assert.Contains(t, output, direction,
				"log output should contain direction")
			assert.Contains(t, output, result,
				"log output should contain result")
		})
	})

	t.Run("FileLogger", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			name := migrationNameGenerator().Draw(t, "migrationName")
			direction := directionGenerator().Draw(t, "direction")
			result := resultGenerator().Draw(t, "result")

			var buf bytes.Buffer
			lg := newFileLoggerWithWriter(LevelDebug, &buf)

			lg.Info("migration %s %s %s", name, direction, result)

			output := buf.String()
			assert.Contains(t, output, name,
				"log output should contain migration name")
			assert.Contains(t, output, direction,
				"log output should contain direction")
			assert.Contains(t, output, result,
				"log output should contain result")
		})
	})
}

// Feature: go-migration, Property 38: Log level filtering
// **Validates: Requirements 18.4**
//
// For any configured log level, only messages at or above that level should be
// output. Messages below the configured level should be suppressed.
func TestProperty38_LogLevelFiltering(t *testing.T) {
	// logAtLevel is a helper that calls the appropriate method on a Logger for
	// the given LogLevel, formatting the message with fmt.Sprintf first.
	logAtLevel := func(lg Logger, lvl LogLevel, msg string) {
		switch lvl {
		case LevelDebug:
			lg.Debug("%s", msg)
		case LevelInfo:
			lg.Info("%s", msg)
		case LevelWarn:
			lg.Warn("%s", msg)
		case LevelError:
			lg.Error("%s", msg)
		}
	}

	t.Run("ConsoleLogger", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			configuredLevel := levelGenerator().Draw(t, "configuredLevel")
			messageLevel := levelGenerator().Draw(t, "messageLevel")
			// Use a unique marker so we can search for it unambiguously.
			marker := rapid.StringMatching(`[a-z]{8,16}`).Draw(t, "marker")

			var buf bytes.Buffer
			lg := newConsoleLoggerWithWriter(configuredLevel, &buf)

			logAtLevel(lg, messageLevel, marker)

			output := buf.String()
			if messageLevel >= configuredLevel {
				assert.Contains(t, output, marker,
					"message at level %s should appear when logger is configured at %s",
					messageLevel, configuredLevel)
				assert.Contains(t, output, fmt.Sprintf("[%s]", messageLevel),
					"output should contain the level tag")
			} else {
				assert.Empty(t, output,
					"message at level %s should be suppressed when logger is configured at %s",
					messageLevel, configuredLevel)
			}
		})
	})

	t.Run("FileLogger", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			configuredLevel := levelGenerator().Draw(t, "configuredLevel")
			messageLevel := levelGenerator().Draw(t, "messageLevel")
			marker := rapid.StringMatching(`[a-z]{8,16}`).Draw(t, "marker")

			var buf bytes.Buffer
			lg := newFileLoggerWithWriter(configuredLevel, &buf)

			logAtLevel(lg, messageLevel, marker)

			output := buf.String()
			if messageLevel >= configuredLevel {
				assert.Contains(t, output, marker,
					"message at level %s should appear when logger is configured at %s",
					messageLevel, configuredLevel)
				assert.Contains(t, output, fmt.Sprintf("[%s]", messageLevel),
					"output should contain the level tag")
			} else {
				assert.Empty(t, output,
					"message at level %s should be suppressed when logger is configured at %s",
					messageLevel, configuredLevel)
			}
		})
	})

	// Verify that multiple messages at mixed levels are correctly filtered.
	t.Run("ConsoleLogger_MultipleMessages", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			configuredLevel := levelGenerator().Draw(t, "configuredLevel")

			var buf bytes.Buffer
			lg := newConsoleLoggerWithWriter(configuredLevel, &buf)

			// Log one message at every level with a unique marker per level.
			markers := make(map[LogLevel]string, len(allLevels))
			for _, lvl := range allLevels {
				m := rapid.StringMatching(`[a-z]{10,16}`).Draw(t, fmt.Sprintf("marker_%s", lvl))
				markers[lvl] = m
				logAtLevel(lg, lvl, m)
			}

			output := buf.String()
			for _, lvl := range allLevels {
				if lvl >= configuredLevel {
					assert.True(t, strings.Contains(output, markers[lvl]),
						"marker for %s should be present (configured=%s)", lvl, configuredLevel)
				} else {
					assert.False(t, strings.Contains(output, markers[lvl]),
						"marker for %s should be absent (configured=%s)", lvl, configuredLevel)
				}
			}
		})
	})
}

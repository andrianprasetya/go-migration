package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- LogLevel tests ---

func TestLogLevelString(t *testing.T) {
	assert.Equal(t, "DEBUG", LevelDebug.String())
	assert.Equal(t, "INFO", LevelInfo.String())
	assert.Equal(t, "WARN", LevelWarn.String())
	assert.Equal(t, "ERROR", LevelError.String())
	assert.Equal(t, "UNKNOWN", LogLevel(99).String())
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"  info  ", LevelInfo},
		{"unknown", LevelInfo}, // default
		{"", LevelInfo},        // default
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseLevel(tt.input))
		})
	}
}

// --- ConsoleLogger tests ---

func TestConsoleLoggerOutputsAtCorrectLevels(t *testing.T) {
	var buf bytes.Buffer
	cl := newConsoleLoggerWithWriter(LevelInfo, &buf)

	cl.Info("hello %s", "world")
	output := buf.String()
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "hello world")
}

func TestConsoleLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	cl := newConsoleLoggerWithWriter(LevelInfo, &buf)

	cl.Debug("should be suppressed")
	assert.Empty(t, buf.String(), "debug message should be suppressed at info level")

	cl.Info("visible info")
	assert.Contains(t, buf.String(), "visible info")

	buf.Reset()
	cl.Warn("visible warn")
	assert.Contains(t, buf.String(), "[WARN]")

	buf.Reset()
	cl.Error("visible error")
	assert.Contains(t, buf.String(), "[ERROR]")
}

func TestConsoleLoggerDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	cl := newConsoleLoggerWithWriter(LevelDebug, &buf)

	cl.Debug("debug msg")
	assert.Contains(t, buf.String(), "[DEBUG]")
	assert.Contains(t, buf.String(), "debug msg")
}

func TestConsoleLoggerErrorLevelSuppressesLower(t *testing.T) {
	var buf bytes.Buffer
	cl := newConsoleLoggerWithWriter(LevelError, &buf)

	cl.Debug("no")
	cl.Info("no")
	cl.Warn("no")
	assert.Empty(t, buf.String())

	cl.Error("yes")
	assert.Contains(t, buf.String(), "[ERROR]")
}

func TestConsoleLoggerFormatIncludesTimestamp(t *testing.T) {
	var buf bytes.Buffer
	cl := newConsoleLoggerWithWriter(LevelDebug, &buf)

	cl.Info("test")
	// RFC3339 timestamps contain "T" and either "Z" or "+"/"-"
	line := buf.String()
	assert.Contains(t, line, "T")
}

// --- FileLogger tests ---

func TestFileLoggerWritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	fl, err := NewFileLogger(path, LevelInfo)
	require.NoError(t, err)
	defer fl.Close()

	fl.Info("file log entry")
	fl.Close()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "file log entry")
	assert.Contains(t, string(data), "[INFO]")
}

func TestFileLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	fl := newFileLoggerWithWriter(LevelWarn, &buf)

	fl.Debug("no")
	fl.Info("no")
	assert.Empty(t, buf.String())

	fl.Warn("yes warn")
	assert.Contains(t, buf.String(), "[WARN]")

	buf.Reset()
	fl.Error("yes error")
	assert.Contains(t, buf.String(), "[ERROR]")
}

func TestFileLoggerInvalidPath(t *testing.T) {
	_, err := NewFileLogger("/nonexistent/dir/file.log", LevelInfo)
	assert.Error(t, err)
}

func TestFileLoggerCloseNilFile(t *testing.T) {
	fl := newFileLoggerWithWriter(LevelInfo, &bytes.Buffer{})
	assert.NoError(t, fl.Close())
}

// --- NopLogger tests ---

func TestNopLoggerDoesNotPanic(t *testing.T) {
	var nop NopLogger
	assert.NotPanics(t, func() {
		nop.Debug("test %d", 1)
		nop.Info("test %d", 2)
		nop.Warn("test %d", 3)
		nop.Error("test %d", 4)
	})
}

// --- Interface compliance ---

func TestConsoleLoggerImplementsLogger(t *testing.T) {
	var _ Logger = (*ConsoleLogger)(nil)
}

func TestFileLoggerImplementsLogger(t *testing.T) {
	var _ Logger = (*FileLogger)(nil)
}

func TestNopLoggerImplementsLogger(t *testing.T) {
	var _ Logger = NopLogger{}
}

// --- formatEntry ---

func TestFormatEntryContainsAllParts(t *testing.T) {
	entry := formatEntry(LevelInfo, "migration %s %s", "users", "up")
	assert.Contains(t, entry, "[INFO]")
	assert.Contains(t, entry, "migration users up")
	assert.True(t, strings.HasSuffix(entry, "\n"))
}

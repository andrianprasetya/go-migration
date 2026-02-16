package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	cmd := newVersionCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "go-migration")
	assert.Contains(t, buf.String(), Version)
	assert.Contains(t, buf.String(), BuildDate)
}

func TestVersionCommandOutput(t *testing.T) {
	// Override version vars for deterministic output.
	origVersion := Version
	origBuild := BuildDate
	Version = "v1.2.3"
	BuildDate = "2024-06-15"
	defer func() {
		Version = origVersion
		BuildDate = origBuild
	}()

	cmd := newVersionCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "go-migration v1.2.3 (built 2024-06-15)\n", buf.String())
}

func TestCommandsNeedingDB(t *testing.T) {
	// Verify the set of commands that require DB access.
	dbCommands := []string{
		"migrate", "migrate:rollback", "migrate:reset",
		"migrate:refresh", "migrate:fresh", "migrate:status",
		"migrate:install", "db:seed",
	}
	for _, name := range dbCommands {
		assert.True(t, commandsNeedingDB[name], "expected %q to need DB", name)
	}

	// These should NOT need DB.
	nonDBCommands := []string{"version", "make:migration", "make:seeder", "help"}
	for _, name := range nonDBCommands {
		assert.False(t, commandsNeedingDB[name], "expected %q to NOT need DB", name)
	}
}

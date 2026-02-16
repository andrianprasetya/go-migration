package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCommand_BasicSetup(t *testing.T) {
	root := NewRootCommand()

	assert.Equal(t, "go-migration", root.Use)
	assert.Contains(t, root.Short, "Laravel-inspired")
	assert.Contains(t, root.Long, "Usage Examples:")
}

func TestNewRootCommand_ConfigFlag(t *testing.T) {
	root := NewRootCommand()

	flag := root.PersistentFlags().Lookup("config")
	require.NotNil(t, flag, "--config flag should be registered")
	assert.Equal(t, "migration.yaml", flag.DefValue)
	assert.Equal(t, "string", flag.Value.Type())
}

func TestNewRootCommand_ConfigFlagCustomValue(t *testing.T) {
	root := NewRootCommand()

	root.SetArgs([]string{"--config", "custom.json"})
	// Execute to parse flags — root has no RunE so it just shows help
	_ = root.Execute()

	val, err := root.PersistentFlags().GetString("config")
	require.NoError(t, err)
	assert.Equal(t, "custom.json", val)
}

func TestNewRootCommand_HelpFlag(t *testing.T) {
	root := NewRootCommand()

	root.SetArgs([]string{"--help"})
	err := root.Execute()

	// --help should not produce an error
	assert.NoError(t, err)
}

func TestNewRootCommand_HelpContainsExamples(t *testing.T) {
	root := NewRootCommand()

	assert.Contains(t, root.Long, "migrate:rollback")
	assert.Contains(t, root.Long, "migrate:status")
	assert.Contains(t, root.Long, "make:migration")
	assert.Contains(t, root.Long, "make:seeder")
	assert.Contains(t, root.Long, "db:seed")
}

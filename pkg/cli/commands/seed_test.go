package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSeedCommand_BasicSetup(t *testing.T) {
	cmd := NewSeedCommand(func() *CommandContext { return nil })
	assert.Equal(t, "db:seed", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewSeedCommand_ClassFlag(t *testing.T) {
	cmd := NewSeedCommand(func() *CommandContext { return nil })
	flag := cmd.Flags().Lookup("class")
	require.NotNil(t, flag, "--class flag should be registered")
	assert.Equal(t, "", flag.DefValue)
	assert.Equal(t, "string", flag.Value.Type())
}

func TestNewSeedCommand_NilContext(t *testing.T) {
	cmd := NewSeedCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seeder runner not initialized")
}

func TestNewSeedCommand_NilSeeder(t *testing.T) {
	cmd := NewSeedCommand(func() *CommandContext { return &CommandContext{} })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seeder runner not initialized")
}

func TestNewSeedCommand_HelpText(t *testing.T) {
	cmd := NewSeedCommand(func() *CommandContext { return nil })
	assert.Contains(t, cmd.Short, "seed")
}

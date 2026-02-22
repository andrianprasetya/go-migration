package commands

import (
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/seeder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSeedRollbackCommand_BasicSetup(t *testing.T) {
	cmd := NewSeedRollbackCommand(func() *CommandContext { return nil })
	assert.Equal(t, "db:seed:rollback", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewSeedRollbackCommand_ClassFlag(t *testing.T) {
	cmd := NewSeedRollbackCommand(func() *CommandContext { return nil })
	flag := cmd.Flags().Lookup("class")
	require.NotNil(t, flag, "--class flag should be registered")
	assert.Equal(t, "", flag.DefValue)
	assert.Equal(t, "string", flag.Value.Type())
}

func TestNewSeedRollbackCommand_NilContext(t *testing.T) {
	cmd := NewSeedRollbackCommand(func() *CommandContext { return nil })
	cmd.Flags().Set("class", "UserSeeder")
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seeder runner not initialized")
}

func TestNewSeedRollbackCommand_NilSeeder(t *testing.T) {
	cmd := NewSeedRollbackCommand(func() *CommandContext { return &CommandContext{} })
	cmd.Flags().Set("class", "UserSeeder")
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seeder runner not initialized")
}

func TestNewSeedRollbackCommand_EmptyClass(t *testing.T) {
	runner := seeder.NewRunner(seeder.NewRegistry(), nil, nil)
	cmd := NewSeedRollbackCommand(func() *CommandContext {
		return &CommandContext{Seeder: runner}
	})
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--class flag is required")
}

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NewMigrateCommand ---

func TestNewMigrateCommand_BasicSetup(t *testing.T) {
	cmd := NewMigrateCommand(func() *CommandContext { return nil })
	assert.Equal(t, "migrate", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMigrateCommand_NilContext(t *testing.T) {
	cmd := NewMigrateCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

func TestNewMigrateCommand_NilMigrator(t *testing.T) {
	cmd := NewMigrateCommand(func() *CommandContext { return &CommandContext{} })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

// --- NewMigrateRollbackCommand ---

func TestNewMigrateRollbackCommand_BasicSetup(t *testing.T) {
	cmd := NewMigrateRollbackCommand(func() *CommandContext { return nil })
	assert.Equal(t, "migrate:rollback", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMigrateRollbackCommand_StepFlag(t *testing.T) {
	cmd := NewMigrateRollbackCommand(func() *CommandContext { return nil })
	flag := cmd.Flags().Lookup("step")
	require.NotNil(t, flag, "--step flag should be registered")
	assert.Equal(t, "0", flag.DefValue)
	assert.Equal(t, "int", flag.Value.Type())
}

func TestNewMigrateRollbackCommand_NilContext(t *testing.T) {
	cmd := NewMigrateRollbackCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

func TestNewMigrateRollbackCommand_NilMigrator(t *testing.T) {
	cmd := NewMigrateRollbackCommand(func() *CommandContext { return &CommandContext{} })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

// --- NewMigrateResetCommand ---

func TestNewMigrateResetCommand_BasicSetup(t *testing.T) {
	cmd := NewMigrateResetCommand(func() *CommandContext { return nil })
	assert.Equal(t, "migrate:reset", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMigrateResetCommand_NilContext(t *testing.T) {
	cmd := NewMigrateResetCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

// --- NewMigrateRefreshCommand ---

func TestNewMigrateRefreshCommand_BasicSetup(t *testing.T) {
	cmd := NewMigrateRefreshCommand(func() *CommandContext { return nil })
	assert.Equal(t, "migrate:refresh", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMigrateRefreshCommand_NilContext(t *testing.T) {
	cmd := NewMigrateRefreshCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

// --- NewMigrateFreshCommand ---

func TestNewMigrateFreshCommand_BasicSetup(t *testing.T) {
	cmd := NewMigrateFreshCommand(func() *CommandContext { return nil })
	assert.Equal(t, "migrate:fresh", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMigrateFreshCommand_NilContext(t *testing.T) {
	cmd := NewMigrateFreshCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

// --- NewMigrateStatusCommand ---

func TestNewMigrateStatusCommand_BasicSetup(t *testing.T) {
	cmd := NewMigrateStatusCommand(func() *CommandContext { return nil })
	assert.Equal(t, "migrate:status", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMigrateStatusCommand_NilContext(t *testing.T) {
	cmd := NewMigrateStatusCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migrator not initialized")
}

// --- NewMigrateInstallCommand ---

func TestNewMigrateInstallCommand_BasicSetup(t *testing.T) {
	cmd := NewMigrateInstallCommand(func() *CommandContext { return nil })
	assert.Equal(t, "migrate:install", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMigrateInstallCommand_NilContext(t *testing.T) {
	cmd := NewMigrateInstallCommand(func() *CommandContext { return nil })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not initialized")
}

func TestNewMigrateInstallCommand_NilDB(t *testing.T) {
	cmd := NewMigrateInstallCommand(func() *CommandContext { return &CommandContext{} })
	err := cmd.RunE(cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not initialized")
}

// --- Help text tests ---

func TestAllMigrateCommands_HaveHelpText(t *testing.T) {
	getCtx := func() *CommandContext { return nil }

	assert.NotEmpty(t, NewMigrateCommand(getCtx).Short)
	assert.NotEmpty(t, NewMigrateRollbackCommand(getCtx).Short)
	assert.NotEmpty(t, NewMigrateResetCommand(getCtx).Short)
	assert.NotEmpty(t, NewMigrateRefreshCommand(getCtx).Short)
	assert.NotEmpty(t, NewMigrateFreshCommand(getCtx).Short)
	assert.NotEmpty(t, NewMigrateStatusCommand(getCtx).Short)
	assert.NotEmpty(t, NewMigrateInstallCommand(getCtx).Short)
}

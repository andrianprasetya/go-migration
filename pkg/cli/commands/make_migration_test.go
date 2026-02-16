package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrianprasetya/go-migration/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMakeMigrationCommand_BasicSetup(t *testing.T) {
	cmd := NewMakeMigrationCommand(func() *CommandContext { return nil })
	assert.Equal(t, "make:migration [name]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMakeMigrationCommand_Flags(t *testing.T) {
	cmd := NewMakeMigrationCommand(func() *CommandContext { return nil })

	createFlag := cmd.Flags().Lookup("create")
	require.NotNil(t, createFlag, "--create flag should be registered")
	assert.Equal(t, "string", createFlag.Value.Type())
	assert.Equal(t, "", createFlag.DefValue)

	tableFlag := cmd.Flags().Lookup("table")
	require.NotNil(t, tableFlag, "--table flag should be registered")
	assert.Equal(t, "string", tableFlag.Value.Type())
	assert.Equal(t, "", tableFlag.DefValue)
}

func TestNewMakeMigrationCommand_NilContext(t *testing.T) {
	cmd := NewMakeMigrationCommand(func() *CommandContext { return nil })
	cmd.SetArgs([]string{"create_users"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generator not initialized")
}

func TestNewMakeMigrationCommand_NilGenerator(t *testing.T) {
	cmd := NewMakeMigrationCommand(func() *CommandContext { return &CommandContext{} })
	cmd.SetArgs([]string{"create_users"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generator not initialized")
}

func TestNewMakeMigrationCommand_MissingArg(t *testing.T) {
	cmd := NewMakeMigrationCommand(func() *CommandContext { return nil })
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestNewMakeMigrationCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	gen := generator.NewGenerator(tmpDir)

	cmd := NewMakeMigrationCommand(func() *CommandContext {
		return &CommandContext{Generator: gen}
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"create_users"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Created migration:")

	// Verify file was actually created
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Contains(t, entries[0].Name(), "create_users.go")
}

func TestNewMakeMigrationCommand_CreateFlag(t *testing.T) {
	tmpDir := t.TempDir()
	gen := generator.NewGenerator(tmpDir)

	cmd := NewMakeMigrationCommand(func() *CommandContext {
		return &CommandContext{Generator: gen}
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"create_users", "--create", "users"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify the generated file contains a Create call
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	content, err := os.ReadFile(filepath.Join(tmpDir, entries[0].Name()))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Create")
	assert.Contains(t, string(content), "users")
}

func TestNewMakeMigrationCommand_TableFlag(t *testing.T) {
	tmpDir := t.TempDir()
	gen := generator.NewGenerator(tmpDir)

	cmd := NewMakeMigrationCommand(func() *CommandContext {
		return &CommandContext{Generator: gen}
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"add_email_to_users", "--table", "users"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify the generated file contains an Alter call
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	content, err := os.ReadFile(filepath.Join(tmpDir, entries[0].Name()))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Alter")
	assert.Contains(t, string(content), "users")
}

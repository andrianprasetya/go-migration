package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/andrianprasetya/go-migration/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMakeSeederCommand_BasicSetup(t *testing.T) {
	cmd := NewMakeSeederCommand(func() *CommandContext { return nil })
	assert.Equal(t, "make:seeder [name]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMakeSeederCommand_NilContext(t *testing.T) {
	cmd := NewMakeSeederCommand(func() *CommandContext { return nil })
	cmd.SetArgs([]string{"user"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generator not initialized")
}

func TestNewMakeSeederCommand_NilGenerator(t *testing.T) {
	cmd := NewMakeSeederCommand(func() *CommandContext { return &CommandContext{} })
	cmd.SetArgs([]string{"user"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generator not initialized")
}

func TestNewMakeSeederCommand_MissingArg(t *testing.T) {
	cmd := NewMakeSeederCommand(func() *CommandContext { return nil })
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestNewMakeSeederCommand_Success(t *testing.T) {
	tmpDir := t.TempDir()
	gen := generator.NewGenerator(tmpDir)

	cmd := NewMakeSeederCommand(func() *CommandContext {
		return &CommandContext{Generator: gen}
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"user"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Created seeder:")

	// Verify file was actually created
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "user_seeder.go", entries[0].Name())
}

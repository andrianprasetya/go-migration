package commands

import (
	"bytes"
	"testing"

	"github.com/andrianprasetya/go-migration/pkg/migrator"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirm_AcceptsY(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("y\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.True(t, result)
	assert.Contains(t, out.String(), "Proceed?")
}

func TestConfirm_AcceptsYes(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("yes\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.True(t, result)
}

func TestConfirm_AcceptsUpperY(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("Y\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.True(t, result)
}

func TestConfirm_AcceptsUpperYES(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("YES\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.True(t, result)
}

func TestConfirm_RejectsN(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("n\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirm_RejectsNo(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("no\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirm_RejectsRandom(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("random\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirm_RejectsEmptyInput(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	result, err := confirm(cmd, "Proceed?")

	require.NoError(t, err)
	assert.False(t, result)
}

func TestConfirm_PromptAppearsInOutput(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString("n\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	_, err := confirm(cmd, "Are you sure?")

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Are you sure?")
	assert.Contains(t, out.String(), "[y/N]")
}

func TestMigrateFresh_ForceSkipsPrompt(t *testing.T) {
	// With --force, the command should skip the confirmation prompt entirely.
	// It will fail because the migrator is nil, but the key assertion is that
	// stdout does NOT contain a confirmation prompt.
	cmd := NewMigrateFreshCommand(func() *CommandContext {
		return &CommandContext{}
	})

	stdin := &bytes.Buffer{} // empty stdin — would block if prompt were shown
	out := &bytes.Buffer{}
	cmd.SetIn(stdin)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--force"})

	_ = cmd.Execute()

	// The prompt should NOT appear because --force was set
	assert.NotContains(t, out.String(), "[y/N]")
}

func TestMigrateFresh_NoForcePromptsAndCancels(t *testing.T) {
	cmd := NewMigrateFreshCommand(func() *CommandContext {
		return &CommandContext{Migrator: migrator.New(nil)}
	})

	stdin := bytes.NewBufferString("n\n")
	out := &bytes.Buffer{}
	cmd.SetIn(stdin)
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "[y/N]")
	assert.Contains(t, out.String(), "Operation cancelled.")
}

func TestMigrateReset_ForceSkipsPrompt(t *testing.T) {
	cmd := NewMigrateResetCommand(func() *CommandContext {
		return &CommandContext{}
	})

	stdin := &bytes.Buffer{}
	out := &bytes.Buffer{}
	cmd.SetIn(stdin)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--force"})

	_ = cmd.Execute()

	assert.NotContains(t, out.String(), "[y/N]")
}

func TestMigrateReset_NoForcePromptsAndCancels(t *testing.T) {
	cmd := NewMigrateResetCommand(func() *CommandContext {
		return &CommandContext{Migrator: migrator.New(nil)}
	})

	stdin := bytes.NewBufferString("n\n")
	out := &bytes.Buffer{}
	cmd.SetIn(stdin)
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "[y/N]")
	assert.Contains(t, out.String(), "Operation cancelled.")
}

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

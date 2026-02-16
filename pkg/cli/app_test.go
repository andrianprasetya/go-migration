package cli

import (
	"testing"

	"github.com/andrianprasetya/go-migration/internal/config"
	"github.com/andrianprasetya/go-migration/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	app := NewApp(nil)

	assert.NotNil(t, app)
	assert.NotNil(t, app.Root())
	assert.Nil(t, app.Context())
}

func TestNewAppWithContext(t *testing.T) {
	ctx := &CLIContext{
		Config: &config.Config{DefaultConn: "test"},
		Logger: logger.NopLogger{},
	}
	app := NewApp(ctx)

	assert.Equal(t, ctx, app.Context())
}

func TestApp_SetContext(t *testing.T) {
	app := NewApp(nil)
	assert.Nil(t, app.Context())

	ctx := &CLIContext{
		Config: &config.Config{DefaultConn: "updated"},
	}
	app.SetContext(ctx)

	assert.Equal(t, ctx, app.Context())
}

func TestApp_RootCommand(t *testing.T) {
	app := NewApp(nil)
	root := app.Root()

	assert.Equal(t, "go-migration", root.Use)
}

func TestApp_RunHelp(t *testing.T) {
	app := NewApp(nil)

	err := app.Run([]string{"--help"})
	assert.NoError(t, err)
}

func TestApp_RunConfigFlag(t *testing.T) {
	app := NewApp(nil)

	err := app.Run([]string{"--config", "custom.yaml"})
	assert.NoError(t, err)

	val, err := app.Root().PersistentFlags().GetString("config")
	require.NoError(t, err)
	assert.Equal(t, "custom.yaml", val)
}

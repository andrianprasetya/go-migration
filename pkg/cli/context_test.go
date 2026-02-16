package cli

import (
	"testing"

	"github.com/andrianprasetya/go-migration/internal/config"
	"github.com/andrianprasetya/go-migration/internal/generator"
	"github.com/andrianprasetya/go-migration/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewCLIContext(t *testing.T) {
	cfg := &config.Config{DefaultConn: "test"}
	gen := generator.NewGenerator("output")
	log := logger.NopLogger{}

	ctx := NewCLIContext(cfg, nil, nil, nil, gen, log)

	assert.NotNil(t, ctx)
	assert.Equal(t, cfg, ctx.Config)
	assert.Nil(t, ctx.DB)
	assert.Nil(t, ctx.Migrator)
	assert.Nil(t, ctx.Seeder)
	assert.Equal(t, gen, ctx.Generator)
	assert.Equal(t, log, ctx.Logger)
}

func TestNewCLIContextAllNil(t *testing.T) {
	ctx := NewCLIContext(nil, nil, nil, nil, nil, nil)

	assert.NotNil(t, ctx)
	assert.Nil(t, ctx.Config)
	assert.Nil(t, ctx.DB)
	assert.Nil(t, ctx.Migrator)
	assert.Nil(t, ctx.Seeder)
	assert.Nil(t, ctx.Generator)
	assert.Nil(t, ctx.Logger)
}

package commands

import (
	"database/sql"

	"github.com/andrianprasetya/go-migration/internal/generator"
	"github.com/andrianprasetya/go-migration/pkg/migrator"
	"github.com/andrianprasetya/go-migration/pkg/seeder"
)

// CommandContext holds the dependencies needed by CLI command handlers.
// It mirrors the fields of cli.CLIContext but lives in the commands package
// to avoid an import cycle between pkg/cli and pkg/cli/commands.
type CommandContext struct {
	DB        *sql.DB
	Migrator  *migrator.Migrator
	Seeder    *seeder.Runner
	Generator *generator.Generator
}

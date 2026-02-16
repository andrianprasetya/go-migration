// Package commands defines the cobra commands for the go-migration CLI.
package commands

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root cobra command for go-migration.
// It sets up the --config persistent flag and help text with usage examples.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "go-migration",
		Short: "Laravel-inspired database migration and seeding tool for Go",
		Long: `go-migration is a Laravel-inspired database migration and seeding tool for Go.

It provides struct-based migrations with up/down methods, a fluent schema builder,
a factory-based seeder system with faker integration, and an intuitive CLI.

Usage Examples:
  go-migration migrate                  Run all pending migrations
  go-migration migrate:rollback         Rollback the last batch of migrations
  go-migration migrate:rollback --step 2  Rollback the last 2 migrations
  go-migration migrate:status           Show migration status
  go-migration migrate:reset            Rollback all migrations
  go-migration migrate:refresh          Reset and re-run all migrations
  go-migration migrate:fresh            Drop all tables and re-run migrations
  go-migration migrate:install          Create the migration tracking table
  go-migration make:migration create_users --create=users
  go-migration make:seeder users
  go-migration db:seed                  Run all seeders
  go-migration db:seed --class=users    Run a specific seeder`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("config", "migration.yaml", "path to configuration file")

	return root
}

package commands

import (
	"fmt"

	"github.com/andrianprasetya/go-migration/pkg/migrator"
	"github.com/spf13/cobra"
)

// NewMigrateInstallCommand creates the "migrate:install" command
// that creates the migration tracking table.
func NewMigrateInstallCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:install",
		Short: "Create the migration tracking table",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.DB == nil {
				return fmt.Errorf("database connection not initialized")
			}
			tracker := migrator.NewTracker(ctx.DB, "migrations")
			return tracker.EnsureTable()
		},
	}
}

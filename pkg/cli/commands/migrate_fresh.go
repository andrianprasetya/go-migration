package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMigrateFreshCommand creates the "migrate:fresh" command
// that drops all tables and re-runs all migrations.
func NewMigrateFreshCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:fresh",
		Short: "Drop all tables and re-run all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}
			return ctx.Migrator.Fresh()
		},
	}
}

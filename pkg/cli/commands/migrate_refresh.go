package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMigrateRefreshCommand creates the "migrate:refresh" command
// that resets all migrations and re-runs them.
func NewMigrateRefreshCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:refresh",
		Short: "Reset and re-run all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}
			return ctx.Migrator.Refresh()
		},
	}
}

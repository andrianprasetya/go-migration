package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMigrateResetCommand creates the "migrate:reset" command that rolls back all migrations.
func NewMigrateResetCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:reset",
		Short: "Rollback all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}
			return ctx.Migrator.Reset()
		},
	}
}

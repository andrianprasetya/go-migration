package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMigrateCommand creates the "migrate" command that runs all pending migrations.
func NewMigrateCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}
			return ctx.Migrator.Up()
		},
	}
}

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMigrateRollbackCommand creates the "migrate:rollback" command.
// It supports a --step flag to roll back a specific number of migrations.
// When --step is 0 (default), it rolls back the last batch.
func NewMigrateRollbackCommand(getCtx func() *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate:rollback",
		Short: "Rollback the last batch of migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}
			steps, err := cmd.Flags().GetInt("step")
			if err != nil {
				return fmt.Errorf("invalid --step flag: %w", err)
			}
			return ctx.Migrator.Rollback(steps)
		},
	}
	cmd.Flags().Int("step", 0, "number of migrations to roll back (0 = last batch)")
	return cmd
}

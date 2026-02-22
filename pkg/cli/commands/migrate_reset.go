package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMigrateResetCommand creates the "migrate:reset" command that rolls back all migrations.
func NewMigrateResetCommand(getCtx func() *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate:reset",
		Short: "Rollback all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				confirmed, err := confirm(cmd, "Are you sure you want to rollback all migrations?")
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Operation cancelled.")
					return nil
				}
			}

			return ctx.Migrator.Reset()
		},
	}

	cmd.Flags().Bool("force", false, "Force the operation to run without confirmation")

	return cmd
}

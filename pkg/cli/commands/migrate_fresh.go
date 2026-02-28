package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMigrateFreshCommand creates the "migrate:fresh" command
// that drops all tables and re-runs all migrations.
func NewMigrateFreshCommand(getCtx func() *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate:fresh",
		Short: "Drop all tables and re-run all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				confirmed, err := confirm(cmd, "Are you sure you want to drop all tables and re-run all migrations?")
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Operation cancelled.")
					return nil
				}
			}

			return ctx.Migrator.Fresh()
		},
	}

	cmd.Flags().Bool("force", false, "Force the operation to run without confirmation")
	cmd.Flags().Bool("dry-run", false, "Show SQL without executing")

	return cmd
}

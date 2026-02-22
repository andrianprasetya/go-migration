package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSeedRollbackCommand creates the "db:seed:rollback" command that rolls back
// a specific seeder by name. The --class flag is required and specifies which
// seeder to roll back.
func NewSeedRollbackCommand(getCtx func() *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db:seed:rollback",
		Short: "Rollback a specific database seeder",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Seeder == nil {
				return fmt.Errorf("seeder runner not initialized")
			}
			class, err := cmd.Flags().GetString("class")
			if err != nil {
				return fmt.Errorf("invalid --class flag: %w", err)
			}
			if class == "" {
				return fmt.Errorf("--class flag is required")
			}
			return ctx.Seeder.Rollback(class)
		},
	}
	cmd.Flags().String("class", "", "seeder class to rollback (required)")
	return cmd
}

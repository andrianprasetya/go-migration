package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSeedTruncateCommand creates the "db:seed:truncate" command that truncates
// (deletes all rows from) a specific seeder table. The --table flag is required
// and specifies which table to truncate.
func NewSeedTruncateCommand(getCtx func() *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db:seed:truncate",
		Short: "Truncate a specific seeder table",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Seeder == nil {
				return fmt.Errorf("seeder runner not initialized")
			}
			table, err := cmd.Flags().GetString("table")
			if err != nil {
				return fmt.Errorf("invalid --table flag: %w", err)
			}
			if table == "" {
				return fmt.Errorf("--table flag is required")
			}
			return ctx.Seeder.Truncate(table)
		},
	}
	cmd.Flags().String("table", "", "table to truncate (required)")
	return cmd
}

package commands

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// NewMigrateStatusCommand creates the "migrate:status" command
// that displays the status of all registered migrations.
func NewMigrateStatusCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:status",
		Short: "Show the status of each migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Migrator == nil {
				return fmt.Errorf("migrator not initialized")
			}
			statuses, err := ctx.Migrator.Status()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "Migration\tStatus\tBatch\tApplied At")
			for _, s := range statuses {
				status := "Pending"
				batch := ""
				appliedAt := ""
				if s.Applied {
					status = "Applied"
					batch = fmt.Sprintf("%d", s.Batch)
					if s.AppliedAt != nil {
						appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, status, batch, appliedAt)
			}
			return w.Flush()
		},
	}
}

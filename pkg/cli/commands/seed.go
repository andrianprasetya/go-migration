package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSeedCommand creates the "db:seed" command that runs database seeders.
// It supports an optional --class flag to run a specific seeder by name.
// When --class is empty, all registered seeders are executed.
func NewSeedCommand(getCtx func() *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db:seed",
		Short: "Run database seeders",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Seeder == nil {
				return fmt.Errorf("seeder runner not initialized")
			}
			class, err := cmd.Flags().GetString("class")
			if err != nil {
				return fmt.Errorf("invalid --class flag: %w", err)
			}
			if class != "" {
				return ctx.Seeder.Run(class)
			}
			return ctx.Seeder.RunAll()
		},
	}
	cmd.Flags().String("class", "", "specific seeder class to run")
	return cmd
}

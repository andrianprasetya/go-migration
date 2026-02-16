package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMakeSeederCommand creates the "make:seeder" command.
// It generates a new seeder file from a template with the correct
// struct scaffolding.
func NewMakeSeederCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "make:seeder [name]",
		Short: "Generate a new seeder file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Generator == nil {
				return fmt.Errorf("generator not initialized")
			}

			path, err := ctx.Generator.Seeder(args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created seeder: %s\n", path)
			return nil
		},
	}
}

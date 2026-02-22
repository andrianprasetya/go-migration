package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewMakeFactoryCommand creates the "make:factory" command.
// It generates a new factory file from a template with the correct
// struct scaffolding.
func NewMakeFactoryCommand(getCtx func() *CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "make:factory [name]",
		Short: "Generate a new factory file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Generator == nil {
				return fmt.Errorf("generator not initialized")
			}

			path, err := ctx.Generator.Factory(args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created factory: %s\n", path)
			return nil
		},
	}
}

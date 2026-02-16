package commands

import (
	"fmt"

	"github.com/andrianprasetya/go-migration/internal/generator"
	"github.com/spf13/cobra"
)

// NewMakeMigrationCommand creates the "make:migration" command.
// It generates a new migration file from a template with the correct
// timestamp prefix and struct scaffolding. Supports --create and --table
// flags for pre-populated schema builder calls.
func NewMakeMigrationCommand(getCtx func() *CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "make:migration [name]",
		Short: "Generate a new migration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getCtx()
			if ctx == nil || ctx.Generator == nil {
				return fmt.Errorf("generator not initialized")
			}

			create, err := cmd.Flags().GetString("create")
			if err != nil {
				return fmt.Errorf("invalid --create flag: %w", err)
			}
			table, err := cmd.Flags().GetString("table")
			if err != nil {
				return fmt.Errorf("invalid --table flag: %w", err)
			}

			opts := generator.MigrationOptions{
				CreateTable: create,
				AlterTable:  table,
			}

			path, err := ctx.Generator.Migration(args[0], opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created migration: %s\n", path)
			return nil
		},
	}
	cmd.Flags().String("create", "", "table name to create (pre-populates schema Create call)")
	cmd.Flags().String("table", "", "table name to alter (pre-populates schema Alter call)")
	return cmd
}

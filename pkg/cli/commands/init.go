package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrianprasetya/go-migration/internal/generator"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the "init" command.
// It scaffolds a new go-migration project with directories, a pre-wired
// cmd/migrator/main.go entry point, and a default go-migration.yaml config file.
// This command does not require a CommandContext or database connection.
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a new go-migration project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return fmt.Errorf("invalid --force flag: %w", err)
			}

			var modulePath string

			if cmd.Flags().Changed("module") {
				modulePath, err = cmd.Flags().GetString("module")
				if err != nil {
					return fmt.Errorf("invalid --module flag: %w", err)
				}
				if modulePath == "" {
					return fmt.Errorf("--module flag value cannot be empty")
				}
			} else {
				goModPath := filepath.Join(cwd, "go.mod")
				modulePath, err = generator.ParseModulePath(goModPath)
				if err != nil {
					return fmt.Errorf("no go.mod found; run 'go mod init' or provide --module flag")
				}
			}

			scaffolder := generator.NewInitScaffolder(cwd, modulePath, force, os.Stdout, os.Stderr)
			_, err = scaffolder.Scaffold()
			return err
		},
	}

	cmd.Flags().String("module", "", "Go module path (overrides go.mod detection)")
	cmd.Flags().Bool("force", false, "Overwrite existing files")

	return cmd
}

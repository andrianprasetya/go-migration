package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version and BuildDate are set at build time via ldflags:
//
//	go build -ldflags "-X main.Version=v1.0.0 -X main.BuildDate=2024-01-01"
var (
	Version   = "dev"
	BuildDate = "unknown"
)

// newVersionCommand creates the "version" sub-command.
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and build date",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "go-migration %s (built %s)\n", Version, BuildDate)
		},
	}
}

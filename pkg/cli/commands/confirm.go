package commands

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// confirm prompts the user with the given message and reads a response from
// cmd.InOrStdin(). It returns true if the user enters "y" or "yes"
// (case-insensitive), and false for any other input.
func confirm(cmd *cobra.Command, message string) (bool, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "%s [y/N]: ", message)

	reader := bufio.NewReader(cmd.InOrStdin())
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input := strings.TrimSpace(strings.ToLower(line))
	return input == "y" || input == "yes", nil
}

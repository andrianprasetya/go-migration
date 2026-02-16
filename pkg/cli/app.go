package cli

import (
	"github.com/andrianprasetya/go-migration/pkg/cli/commands"
	"github.com/spf13/cobra"
)

// App is the top-level CLI application. It holds the root cobra command
// and the CLIContext that provides dependencies to command handlers.
type App struct {
	root *cobra.Command
	ctx  *CLIContext
}

// NewApp creates a new CLI App. The CLIContext may be nil during initial
// setup — it is wired after config loading in the entry point (task 14.5).
func NewApp(ctx *CLIContext) *App {
	root := commands.NewRootCommand()
	return &App{
		root: root,
		ctx:  ctx,
	}
}

// Root returns the root cobra command, allowing callers to add sub-commands
// or inspect the command tree.
func (a *App) Root() *cobra.Command {
	return a.root
}

// Context returns the CLIContext.
func (a *App) Context() *CLIContext {
	return a.ctx
}

// SetContext sets the CLIContext. This is used when the context is built
// after config loading but before command execution.
func (a *App) SetContext(ctx *CLIContext) {
	a.ctx = ctx
}

// Run executes the root command with the given arguments.
// Pass nil to use os.Args[1:].
func (a *App) Run(args []string) error {
	if args != nil {
		a.root.SetArgs(args)
	}
	return a.root.Execute()
}

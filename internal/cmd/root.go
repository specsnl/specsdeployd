// Package cmd is the specsdeployd command tree: one file per command, all of
// them leaves on the root built in [NewRootCmd].
//
// Commands return errors; main turns them into an exit code. Nothing here calls
// os.Exit — it skips deferred cleanup.
package cmd

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// AppName is how the binary refers to itself in its own output.
const AppName = "specsdeployd"

const flagVersion = "version"

// App holds what every command needs beyond its own flags. Commands close over
// it, so a test can point the streams at buffers.
type App struct {
	// Out carries a command's product; Err carries a failure, and outlives the
	// run because main prints through it after Execute returns.
	Out io.Writer
	Err io.Writer
}

// NewApp creates an App over the real process streams. Both are replaced in
// PersistentPreRun by the running command's own streams.
func NewApp() *App {
	return &App{
		Out: os.Stdout,
		Err: os.Stderr,
	}
}

// Execute builds the command tree and runs it with a background context.
func Execute(app *App) error {
	return ExecuteContext(context.Background(), app)
}

// ExecuteContext builds the command tree and runs it with the given context.
func ExecuteContext(ctx context.Context, app *App) error {
	return NewRootCmd(app).ExecuteContext(ctx)
}

// NewRootCmd builds the root command with every subcommand attached. Exported
// so a test can drive the tree over its own streams.
func NewRootCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   AppName,
		Short: "The deploy agent of the Specs golden images",
		Long: `specsdeployd is the deploy agent of the Specs golden images: a host systemd
service that receives GitHub deployment webhooks and rolls the target
application forward.

None of that is implemented yet. This binary answers "version" and no more.

Use "specsdeployd <command> --help" for more information about a command.`,

		// Usage after a runtime failure buries the line that matters, and
		// Cobra's own error print would duplicate main's.
		SilenceUsage:  true,
		SilenceErrors: true,

		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			app.Out = cmd.OutOrStdout()
			app.Err = cmd.ErrOrStderr()
		},

		// Args stays nil so Cobra keeps rejecting unknown subcommands instead of
		// showing this help for them.
		RunE: func(cmd *cobra.Command, _ []string) error {
			if version, _ := cmd.Flags().GetBool(flagVersion); version {
				writeVersion(app, true)

				return nil
			}

			return cmd.Help()
		},
	}

	// Not cmd.Version: Cobra handles its built-in version flag before
	// PersistentPreRun, so the value would bypass the command's streams.
	cmd.Flags().Bool(flagVersion, false, "Print the bare version and exit")

	cmd.AddCommand(
		newVersionCmd(app),
	)

	return cmd
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the binary version, injected at build time:
//
//	-ldflags "-X github.com/specsnl/specsdeployd/internal/cmd.Version=1.2.3"
//
// .goreleaser.yml and the Dockerfile name this variable by that exact path, in
// a string no compiler checks. Rename or move it and every release silently
// ships as "dev".
var Version = "dev"

func newVersionCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the specsdeployd version",
		Long: `Print the specsdeployd version.

The version is a result, not narration, so it goes to stdout and can be
captured:

  specsdeployd version     # specsdeployd version 1.2.3
  specsdeployd --version   # 1.2.3`,
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			writeVersion(app, false)

			return nil
		},
	}
}

// writeVersion is shared with the root's --version so the two renderings of the
// same variable cannot drift. bare drops the surrounding sentence, for
// $(specsdeployd --version).
func writeVersion(app *App, bare bool) {
	format := AppName + " version %s\n"
	if bare {
		format = "%s\n"
	}

	fmt.Fprintf(app.Out, format, Version)
}

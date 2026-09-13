// Command specsdeployd is the deploy agent of the Specs golden images. It
// currently answers "version" and no more.
package main

import (
	"fmt"
	"os"

	"github.com/specsnl/specsdeployd/internal/cmd"
)

func main() {
	app := cmd.NewApp()

	if err := cmd.Execute(app); err != nil {
		fmt.Fprintln(app.Err, "Error:", err)

		os.Exit(1)
	}
}

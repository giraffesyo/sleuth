package cli

import (
	"github.com/giraffesyo/sleuth/internal/cli/search"
	"github.com/spf13/cobra"
)

// Set by the --verbose flag
var VerboseLogging bool

var (
	use   = "sleuth"
	short = "The Sleuth CLI"
)

var RootCmd = &cobra.Command{
	Use:               use,
	Short:             short,
	Version:           "v0.0.1",
	DisableAutoGenTag: true,
}

func init() {
	RootCmd.AddCommand(search.Cmd)
}

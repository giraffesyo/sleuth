package cli

import (
	"os"

	"github.com/giraffesyo/sleuth/internal/cli/aicheck"
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
	Run:               run,
}

func run(cmd *cobra.Command, args []string) {

	os.Stderr.WriteString(`
         .-""-.    ____  _     _____ _   _ _____ _   _ 
 _______/      \  / ___|| |   | ____| | | |_   _| | | |
|_______        ; \___ \| |   |  _| | | | | | | | |_| |
        \      /   ___) | |___| |___| |_| | | | |  _  |
         '-..-'   |____/|_____|_____|\___/  |_| |_| |_|
		`)
	cmd.Help()
}

func init() {
	RootCmd.AddCommand(search.Cmd)
	RootCmd.AddCommand(aicheck.Cmd)
}

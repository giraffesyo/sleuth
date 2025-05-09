package cli

import (
	"os"

	"github.com/giraffesyo/sleuth/internal/cli/aicheck"
	"github.com/giraffesyo/sleuth/internal/cli/csv"
	determineLocation "github.com/giraffesyo/sleuth/internal/cli/determine_location"
	determineVictim "github.com/giraffesyo/sleuth/internal/cli/determine_victim"
	downloadVideos "github.com/giraffesyo/sleuth/internal/cli/download_videos"
	generateQueries "github.com/giraffesyo/sleuth/internal/cli/generate_queries"
	ingestTimestamps "github.com/giraffesyo/sleuth/internal/cli/ingest_timestamps"
	"github.com/giraffesyo/sleuth/internal/cli/search"
	showQueries "github.com/giraffesyo/sleuth/internal/cli/show_queries"
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
	RootCmd.AddCommand(csv.Cmd)
	RootCmd.AddCommand(downloadVideos.Cmd)
	RootCmd.AddCommand(ingestTimestamps.Cmd)
	RootCmd.AddCommand(generateQueries.Cmd)
	RootCmd.AddCommand(determineVictim.Cmd)
	RootCmd.AddCommand(determineLocation.Cmd)
	RootCmd.AddCommand(showQueries.Cmd)
}

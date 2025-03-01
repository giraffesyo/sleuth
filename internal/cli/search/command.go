package search

import (
	"github.com/giraffesyo/sleuth/internal/sleuth"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var query string

var (
	use   = "search"
	short = "Search for news articles with the provided term"
)

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	sleuth := sleuth.NewSleuth(
		sleuth.WithProvider(sleuth.ProviderCNN),
		sleuth.WithProvider(sleuth.ProviderFoxNews),
		sleuth.WithSearchQuery(query),
	)
	err := sleuth.Run()
	if err != nil {
		log.Err(err).Msg("failed to run sleuth")
	}
}

func init() {
	Cmd.Flags().StringVarP(&query, "query", "q", "", "The search terms to use, wrap multiple words in quotes")
	Cmd.MarkFlagRequired("query")
}

package search

import (
	"fmt"

	"github.com/giraffesyo/sleuth/internal/sleuth"
	"github.com/giraffesyo/sleuth/internal/sleuth/providers/cnn"
	"github.com/giraffesyo/sleuth/internal/sleuth/providers/fox"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var defaultProviders = []string{cnn.ProviderCNN, fox.ProviderFoxNews}
var query string
var enabledProviders []string

var (
	use   = "search"
	short = "Search for news articles with the provided term"
)

func long() string {
	longHelp := `
Search for news articles with the provided term. The search term must be provided.

You can specify the providers to use for searching, if not provided all providers will be used.

Providers are:`

	for _, p := range defaultProviders {
		longHelp += fmt.Sprintf("\n- %s", p)
	}
	return longHelp
}

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Run:   run,
	Long:  long(),
}

func run(cmd *cobra.Command, args []string) {
	sleuth := sleuth.NewSleuth(
		sleuth.WithProvider(enabledProviders...),
		sleuth.WithSearchQuery(query),
	)
	err := sleuth.Run()
	if err != nil {
		log.Err(err).Msg("failed to run sleuth")
	}
}

func init() {
	Cmd.Flags().StringVarP(&query, "query", "q", "", "The search terms to use, wrap multiple words in quotes")
	Cmd.Flags().StringSliceVarP(&enabledProviders, "providers", "p", defaultProviders, "The providers to use for searching, if not provided all providers will be used")
	Cmd.MarkFlagRequired("query")
}

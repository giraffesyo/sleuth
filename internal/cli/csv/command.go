package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var outputFile string

var (
	use   = "csv"
	short = "Export articles to CSV format"
)

func long() string {
	return "Export articles from the database to CSV format. By default outputs to stdout, but can write to a file with -o flag."
}

// timestampsToJSON marshals the entire slice into a JSON array string.
func timestampsToJSON(timestamps []db.RelevantTimestamp) string {
	b, err := json.Marshal(timestamps)
	if err != nil {
		// in case of error, fall back to empty array
		return "[]"
	}
	return string(b)
}

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Long:  long(),
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	uri := db.GetMongoURI()
	if err := db.Models.ConnectDatabase(uri); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Fetch all articles from the database
	articles, err := db.Models.FindAllArticles(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to fetch articles")
	}

	if len(articles) == 0 {
		log.Info().Msg("no articles found in database")
		return
	}

	log.Info().Int("count", len(articles)).Msg("found articles")

	// Determine where to write the CSV data
	var output *os.File
	if outputFile != "" {
		output, err = os.Create(outputFile)
		if err != nil {
			log.Fatal().Err(err).Str("file", outputFile).Msg("failed to create output file")
		}
		defer output.Close()
		log.Info().Str("file", outputFile).Msg("writing CSV to file")
	} else {
		output = os.Stdout
		log.Info().Msg("writing CSV to stdout")
	}

	// Create CSV writer
	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write header row
	header := []string{
		"ID", "Title", "URL", "Date", "Description", "Provider",
		"AI Checked", "AI Suggests Download", "Victim Names", "Location",
		"Case ID", "Relevant Timestamps",
	}
	if err := writer.Write(header); err != nil {
		log.Fatal().Err(err).Msg("error writing CSV header")
	}

	// Write article data
	for _, article := range articles {
		row := []string{
			article.Id.Hex(),
			article.Title,
			article.Url,
			article.Date,
			article.Description,
			article.Provider,
			fmt.Sprintf("%t", article.AiHasCheckedIfShouldDownloadVideo),
			fmt.Sprintf("%t", article.AiSuggestsDownloadingVideo),
			strings.Join(article.VictimNames, ", "),
			article.Location,
			fmt.Sprintf("%d", article.CaseId),
			timestampsToJSON(article.RelevantTimestamps),
		}
		if err := writer.Write(row); err != nil {
			log.Error().Err(err).Str("url", article.Url).Msg("error writing article to CSV")
			continue
		}
	}

	if outputFile != "" {
		log.Info().Str("file", outputFile).Int("articles", len(articles)).Msg("CSV export complete")
	}
}

func init() {
	Cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (if not provided, outputs to stdout)")
}

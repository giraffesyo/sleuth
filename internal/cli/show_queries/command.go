package showqueries

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	use   = "show-queries"
	short = "Display all search queries stored in the database"
	long  = "Display all search queries that have been generated and stored in the database, including whether they have been used in searches."

	// Command flags
	outputFile string
	jsonFormat bool
	onlyUnused bool
)

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Long:  long,
	Run:   run,
}

func init() {
	Cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (if not specified, outputs to stdout)")
	Cmd.Flags().BoolVarP(&jsonFormat, "json", "j", false, "Output in JSON format")
	Cmd.Flags().BoolVarP(&onlyUnused, "unused", "u", false, "Show only unused queries")
}

func run(cmd *cobra.Command, args []string) {
	// Initialize database connection
	uri := db.GetMongoURI()
	if err := db.Models.ConnectDatabase(uri); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
		return
	}

	ctx := context.Background()

	var queries []*db.Query
	var err error

	// Retrieve queries based on the flags
	if onlyUnused {
		log.Info().Msg("retrieving unused queries from database")
		queries, err = db.Models.FindUnusedQueries(ctx)
	} else {
		log.Info().Msg("retrieving all queries from database")
		queries, err = db.Models.FindAllQueries(ctx)
	}

	if err != nil {
		log.Fatal().Err(err).Msg("failed to retrieve queries")
		return
	}

	log.Info().Int("count", len(queries)).Msg("queries found")

	// Prepare output
	var output string
	if jsonFormat {
		// JSON output format
		jsonData, err := json.MarshalIndent(queries, "", "  ")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to marshal queries to JSON")
			return
		}
		output = string(jsonData)
	} else {
		// Plain text output format
		output = fmt.Sprintf("Found %d queries:\n\n", len(queries))
		for i, q := range queries {
			usedStatus := "✓ Used"
			if !q.Used {
				usedStatus = "✗ Unused"
			}

			output += fmt.Sprintf("[%d][%s] %s", i+1, usedStatus, q.Query)

			output += "\n"
		}
	}

	// Output to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			log.Fatal().Err(err).Str("file", outputFile).Msg("failed to write to output file")
			return
		}
		log.Info().Str("file", outputFile).Msg("queries written to file")
	} else {
		fmt.Println(output)
	}
}

package ingest_timestamps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	use   = "ingest-timestamp-metadata"
	short = "Ingest timestamp metadata from JSON files and update MongoDB documents"
	// Directory where timestamp files are stored
	timestampsDir = "./timestamps"
)

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	uri := db.GetMongoURI()
	if err := db.Models.ConnectDatabase(uri); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Check if timestamps directory exists
	if _, err := os.Stat(timestampsDir); os.IsNotExist(err) {
		log.Fatal().Str("dir", timestampsDir).Msg("timestamps directory does not exist")
	}

	// Read all files from the timestamps directory
	files, err := os.ReadDir(timestampsDir)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read timestamps directory")
	}

	if len(files) == 0 {
		log.Info().Msg("no timestamp files found")
		return
	}

	log.Info().Int("count", len(files)).Msg("found timestamp files to process")

	// Process each file
	successCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		filename := file.Name()
		// Check if the filename has a JSON extension or skip it otherwise
		if !strings.HasSuffix(strings.ToLower(filename), ".json") {
			log.Warn().Str("file", filename).Msg("skipping non-JSON file")
			continue
		}

		// Extract the ObjectID from the filename (remove the .json extension)
		idStr := strings.TrimSuffix(filename, filepath.Ext(filename))
		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			log.Warn().Str("file", filename).Err(err).Msg("invalid ObjectID in filename")
			continue
		}

		// Read and parse the JSON file
		filePath := filepath.Join(timestampsDir, filename)
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Str("file", filename).Err(err).Msg("failed to read file")
			continue
		}

		// Validate JSON before updating
		var timestampData interface{}
		if err := json.Unmarshal(fileBytes, &timestampData); err != nil {
			log.Error().Str("file", filename).Err(err).Msg("invalid JSON in file")
			continue
		}

		// Update the MongoDB document
		update := bson.M{
			"relevantTimestamps": timestampData,
		}
		err = db.Models.UpdateArticle(ctx, objID, update)
		if err != nil {
			log.Error().Str("file", filename).Str("objectId", objID.Hex()).Err(err).Msg("failed to update article")
			continue
		}

		log.Info().Str("id", objID.Hex()).Msg("successfully updated article with timestamp metadata")
		successCount++
	}

	log.Info().Int("total", len(files)).Int("success", successCount).Msg("timestamp ingestion completed")
}

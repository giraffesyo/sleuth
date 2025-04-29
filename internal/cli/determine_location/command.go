package determinelocation

import (
	"fmt"
	"strings"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	use   = "determine-location"
	short = "Analyzes articles to determine the location where bodies were found"

	// Command flags
	modelName string
	limit     int
)

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Run:   run,
}

func init() {
	// Add flags to the command
	Cmd.Flags().StringVarP(&modelName, "model", "m", "llama3.1", "Model to use for determining location")
	Cmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit the number of articles to process (0 means no limit)")
}

func run(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	uri := db.GetMongoURI()
	if err := db.Models.ConnectDatabase(uri); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Find articles where location field is not set
	filter := bson.M{"location": bson.M{"$exists": false}}
	articles, err := db.Models.FindArticlesByFilter(ctx, filter)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find articles without location")
	}

	if len(articles) == 0 {
		log.Info().Msg("no articles found without location information")
		return
	}

	count := len(articles)
	if limit > 0 && limit < count {
		count = limit
		articles = articles[:limit]
	}

	log.Info().Int("count", count).Msg("found articles that need location determination")

	for i, article := range articles {
		log.Info().Int("current", i+1).Int("total", count).Str("url", article.Url).Msg("determining location")

		// Combine all available data about the article to help determine the location
		fullPrompt := fmt.Sprintf("Title: %s\nDescription: %s\nDate: %s\nProvider: %s\n",
			article.Title,
			article.Description,
			article.Date,
			article.Provider)

		// Determine the location using Ollama LLM
		location, err := determineLocation(fullPrompt)
		if err != nil {
			log.Err(err).Str("url", article.Url).Msg("failed to determine location, skipping")
			continue
		}

		// Update the article with the determined location
		update := bson.M{
			"location": location,
		}

		err = db.Models.UpdateArticle(ctx, article.Id, update)
		if err != nil {
			log.Err(err).Str("url", article.Url).Msg("failed to update article with location")
			continue
		}

		log.Info().Str("url", article.Url).Str("location", location).Msg("updated article with location")
	}

	log.Info().Msg("finished determining locations")
}

// determineLocation calls the Ollama API to determine the location from the article data
func determineLocation(articleData string) (string, error) {
	systemPrompt := `You are an AI helping to identify locations in news articles about missing persons and bodies found.
Based on the article information provided, determine the location where the body was found or the incident occurred.
Return ONLY the location name with no explanations or additional text.
Be as specific as possible, including city, state, country, or other geographical indicators if available.
If you cannot determine a location, respond with "Unknown".
Format your response as a simple location string, for example: "Denver, Colorado" or "Lake Michigan near Chicago".
`
	// Call local Ollama API
	response, err := CallOllama(modelName, systemPrompt, articleData)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}

	// Clean up the response
	return cleanResponse(response), nil
}

// cleanResponse cleans up the LLM response to ensure it's a usable format
func cleanResponse(response string) string {
	// Remove quotes and trim whitespace
	cleaned := strings.TrimSpace(response)

	// Strip any leading/trailing quotes
	if len(cleaned) > 0 && (cleaned[0] == '"' || cleaned[0] == '\'') {
		cleaned = cleaned[1:]
	}
	if len(cleaned) > 0 && (cleaned[len(cleaned)-1] == '"' || cleaned[len(cleaned)-1] == '\'') {
		cleaned = cleaned[:len(cleaned)-1]
	}

	// Remove any "location:", "place:" prefixes
	prefixes := []string{"location:", "place:", "location is:", "location at:", "found at:", "found in:"}
	for _, prefix := range prefixes {
		if len(cleaned) > len(prefix) && startsCaseInsensitive(cleaned, prefix) {
			cleaned = cleaned[len(prefix):]
			cleaned = strings.TrimSpace(cleaned)
		}
	}

	// If empty after cleaning, return Unknown
	if cleaned == "" {
		return "Unknown"
	}

	return cleaned
}

// startsCaseInsensitive checks if a string starts with a prefix (case insensitive)
func startsCaseInsensitive(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		if lowerChar(s[i]) != lowerChar(prefix[i]) {
			return false
		}
	}
	return true
}

// lowerChar converts a byte to lowercase if it's an ASCII letter
func lowerChar(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

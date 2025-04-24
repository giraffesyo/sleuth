package determinevictim

import (
	"fmt"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	use   = "determine-victim"
	short = "Analyzes articles to determine the victim's name"

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
	Cmd.Flags().StringVarP(&modelName, "model", "m", "llama3.1", "Model to use for determining victim name")
	Cmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit the number of articles to process (0 means no limit)")
}

func run(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	uri := db.GetMongoURI()
	if err := db.Models.ConnectDatabase(uri); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Find articles where name is not set
	// The query looks for articles where victimName field doesn't exist
	filter := bson.M{"victimName": bson.M{"$exists": false}}
	articles, err := db.Models.FindArticlesByFilter(ctx, filter)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find articles without victim names")
	}

	if len(articles) == 0 {
		log.Info().Msg("no articles found without victim names")
		return
	}

	count := len(articles)
	if limit > 0 && limit < count {
		count = limit
		articles = articles[:limit]
	}

	log.Info().Int("count", count).Msg("found articles that need victim name determination")

	for i, article := range articles {
		log.Info().Int("current", i+1).Int("total", count).Str("url", article.Url).Msg("determining victim name")

		// Combine all available data about the article to help determine the victim
		fullPrompt := fmt.Sprintf("Title: %s\nDescription: %s\nDate: %s\nProvider: %s\n",
			article.Title,
			article.Description,
			article.Date,
			article.Provider)

		// Determine the victim name using Ollama LLM
		victimName, err := determineVictimName(fullPrompt)
		if err != nil {
			log.Err(err).Str("url", article.Url).Msg("failed to determine victim name, skipping")
			continue
		}

		// Update the article with the determined victim name
		update := bson.M{
			"victimName": victimName,
		}

		err = db.Models.UpdateArticle(ctx, article.Id, update)
		if err != nil {
			log.Err(err).Str("url", article.Url).Msg("failed to update article with victim name")
			continue
		}

		log.Info().Str("url", article.Url).Str("victimName", victimName).Msg("updated article with victim name")
	}

	log.Info().Msg("finished determining victim names")
}

// determineVictimName calls the Ollama API to determine the victim's name from the article data
func determineVictimName(articleData string) (string, error) {
	systemPrompt := `You are an AI helping to identify victims in news articles about missing persons and bodies found.
Based on the article information provided, determine the name of the victim.
If multiple people are mentioned, identify who is the actual victim (the missing or deceased person).
Return ONLY the victim's full name with no explanations or additional text.
If you cannot determine the victim's name, respond with "Unknown".
Do not include titles (Mr., Mrs., Dr., etc.) unless they are part of a formal name like "Dr. Martin Luther King Jr.".
`
	// Call local Ollama API
	response, err := CallOllama(modelName, systemPrompt, articleData)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}

	// Clean up the response (remove quotes, extra whitespace, etc.)
	cleanedResponse := cleanResponse(response)

	// Validate the response
	if len(cleanedResponse) < 2 || len(cleanedResponse) > 100 {
		return "Unknown", nil
	}

	return cleanedResponse, nil
}

// cleanResponse cleans up the LLM response to ensure it's a usable name
func cleanResponse(response string) string {
	// Remove quotes and trim whitespace
	cleaned := response

	// Strip any leading/trailing quotes
	if len(cleaned) > 0 && (cleaned[0] == '"' || cleaned[0] == '\'') {
		cleaned = cleaned[1:]
	}
	if len(cleaned) > 0 && (cleaned[len(cleaned)-1] == '"' || cleaned[len(cleaned)-1] == '\'') {
		cleaned = cleaned[:len(cleaned)-1]
	}

	// Remove any "name:", "victim:" prefixes
	prefixes := []string{"name:", "victim:", "victim name:", "name is:", "victim is:"}
	for _, prefix := range prefixes {
		if len(cleaned) > len(prefix) && startsCaseInsensitive(cleaned, prefix) {
			cleaned = cleaned[len(prefix):]
		}
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

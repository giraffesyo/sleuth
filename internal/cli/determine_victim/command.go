package determinevictim

import (
	"fmt"
	"strings"

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

	// Find articles where victim names array is not set
	// The query looks for articles where victimNames field doesn't exist
	filter := bson.M{"victimNames": bson.M{"$exists": false}}
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
		log.Info().Int("current", i+1).Int("total", count).Str("url", article.Url).Msg("determining victim names")

		// Combine all available data about the article to help determine the victims
		fullPrompt := fmt.Sprintf("Title: %s\nDescription: %s\nDate: %s\nProvider: %s\n",
			article.Title,
			article.Description,
			article.Date,
			article.Provider)

		// Determine the victim names using Ollama LLM
		victimNames, err := determineVictimNames(fullPrompt)
		if err != nil {
			log.Err(err).Str("url", article.Url).Msg("failed to determine victim names, skipping")
			continue
		}

		// Update the article with the determined victim names
		update := bson.M{
			"victimNames": victimNames,
		}

		err = db.Models.UpdateArticle(ctx, article.Id, update)
		if err != nil {
			log.Err(err).Str("url", article.Url).Msg("failed to update article with victim names")
			continue
		}

		log.Info().Str("url", article.Url).Strs("victimNames", victimNames).Msg("updated article with victim names")
	}

	log.Info().Msg("finished determining victim names")
}

// determineVictimNames calls the Ollama API to determine the victim names from the article data
func determineVictimNames(articleData string) ([]string, error) {
	systemPrompt := `You are an AI helping to identify victims in news articles about missing persons and bodies found.
Based on the article information provided, determine the name(s) of the victim(s).
If multiple victims are mentioned, return all of their names separated by semicolons (;).
Return ONLY the victim names with no explanations or additional text.
If you cannot determine any victim's name, respond with "Unknown".
Do not include titles (Mr., Mrs., Dr., etc.) unless they are part of a formal name like "Dr. Martin Luther King Jr.".
Format your response as: "Name1; Name2; Name3" if multiple victims are present.
`
	// Call local Ollama API
	response, err := CallOllama(modelName, systemPrompt, articleData)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama: %w", err)
	}

	// Clean up the response and split into multiple names if needed
	return parseVictimNames(response), nil
}

// parseVictimNames processes the LLM response and returns a list of victim names
func parseVictimNames(response string) []string {
	// Clean up the response (remove quotes, extra whitespace, etc.)
	cleaned := cleanResponse(response)

	// Handle the case where no victims were found
	if cleaned == "" || cleaned == "Unknown" || cleaned == "unknown" {
		return []string{"Unknown"}
	}

	// Split by semicolons, which is our requested delimiter for multiple victims
	names := strings.Split(cleaned, ";")

	// Clean each name and filter out empty ones
	var validNames []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" && len(name) >= 2 {
			validNames = append(validNames, name)
		}
	}

	// If we didn't find any valid names, return Unknown
	if len(validNames) == 0 {
		return []string{"Unknown"}
	}

	return validNames
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

	// Remove any "name:", "victim:" prefixes
	prefixes := []string{"name:", "victim:", "victim name:", "name is:", "victim is:", "victims:", "victim names:"}
	for _, prefix := range prefixes {
		if len(cleaned) > len(prefix) && startsCaseInsensitive(cleaned, prefix) {
			cleaned = cleaned[len(prefix):]
			cleaned = strings.TrimSpace(cleaned)
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

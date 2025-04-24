package generatequeries

import (
	"fmt"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	use   = "generate-queries"
	short = "Generates search queries for finding news articles about missing persons and bodies found"

	// Command flags
	customPrompt string
	numQueries   int
)

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Run:   run,
}

func init() {
	// Add flags to the command
	Cmd.Flags().StringVarP(&customPrompt, "prompt", "p", "", "Custom prompt to use for query generation")
	Cmd.Flags().IntVarP(&numQueries, "num", "n", 5, "Number of queries to generate (default 5)")
}

func run(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	uri := db.GetMongoURI()
	if err := db.Models.ConnectDatabase(uri); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	// Use default prompt if no custom prompt is provided
	prompt := getPrompt()
	if customPrompt != "" {
		prompt = customPrompt
	}

	log.Info().Msg("Generating search queries...")

	for i := 0; i < numQueries; i++ {
		log.Info().Int("query", i+1).Int("total", numQueries).Msg("Generating query")

		// Generate query using Ollama LLM
		queryString, err := generateQuery(prompt)
		if err != nil {
			log.Err(err).Msg("failed to generate query, trying again")
			continue
		}

		// Check if this query already exists to avoid duplicates
		existingQuery, err := db.Models.FindQueryByValue(ctx, queryString)
		if err == nil && existingQuery != nil {
			log.Info().Str("query", queryString).Msg("Query already exists, skipping")
			continue
		}

		// Create a new query
		query := &db.Query{
			Query:       queryString,
			Used:        false,
			Description: fmt.Sprintf("Generated using prompt: %s", prompt),
		}

		// Save to database
		err = db.Models.CreateQuery(ctx, query)
		if err != nil {
			log.Err(err).Str("query", queryString).Msg("failed to save query")
			continue
		}

		log.Info().Str("query", queryString).Msg("Successfully generated and saved query")
	}
}

// getPrompt returns the default prompt for generating search queries
func getPrompt() string {
	return `Generate a single search query that would be effective for finding news articles about missing persons cases or cases where bodies have been found. 
The query should be specific enough to return relevant results but general enough to capture a wide range of cases. 
Focus on creating search terms that would help build a dataset for tracking missing persons and bodies found.
Return only the search query string with no explanations or additional text.
Example formats: "body found in lake", "remains discovered woods", "missing person case solved".`
}

// generateQuery calls the Ollama API to generate a search query
func generateQuery(prompt string) (string, error) {
	systemPrompt := `You are helping to build a dataset of news articles about missing persons and bodies found. 
Generate only a single search query that would help find relevant news articles.
Return ONLY the search query with no explanations, quotes, or additional text.
Do not use any special characters or database syntax.
Do not join multiple queries together in any way.
Do not attempt to return more than one query at a time.
Do not use conjunctions (and, or, +, etc.) to combine multiple queries.
`

	// Call local Ollama API
	response, err := CallOllama("llama3.1", systemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}

	// Clean up the response (remove quotes, extra whitespace, etc.)
	cleanedResponse := cleanResponse(response)

	// Validate the response
	if len(cleanedResponse) < 3 || len(cleanedResponse) > 255 {
		return "", fmt.Errorf("invalid query length: %d characters", len(cleanedResponse))
	}

	return cleanedResponse, nil
}

// cleanResponse cleans up the LLM response to ensure it's a usable search query
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

	// Remove any "query:", "search:" prefixes
	prefixes := []string{"query:", "search:", "search query:", "q:"}
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

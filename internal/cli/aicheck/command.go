package aicheck

import (
	"encoding/json"
	"fmt"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	use   = "aicheck"
	short = "Goes through all articles and decides if they should be downloaded"
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

	articles, err := db.Models.FindAllArticlesNotChecked(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find articles")
	}
	if len(articles) == 0 {
		log.Info().Msg("no articles to check")
		return
	}

	log.Info().Int("count", len(articles)).Msg("found articles to check")
	for _, article := range articles {
		log.Info().Str("url", article.Url).Msg("checking article")
		shouldDownload, err := decideIfShouldDownloadVideo(article)
		if err != nil {
			log.Err(err).Msg("failed to decide if should download video, skipping for now")
			continue
		}
		log.Info().Str("url", article.Url).Bool("shouldDownload", shouldDownload).Msg("decided if should download video")
		// set aiHasCheckedIfShouldDownloadVideo to true and shouldDownload to the value we got
		update := bson.M{
			"aiHasCheckedIfShouldDownloadVideo": true,
			"aiSuggestsDownloadingVideo":        shouldDownload,
		}
		err = db.Models.UpdateArticle(ctx, article.Id, update)
		if err != nil {
			log.Err(err).Msg("failed to update article")
			continue
		}
		log.Info().Str("url", article.Url).Msg("updated article")

	}
}

func decideIfShouldDownloadVideo(article *db.Article) (bool, error) {

	systemPrompt := `We are building a dataset on crime cases where bodies were found. I will provide you with a video title and description and you will decide if the video should be downloaded for further processing.

A video would be useful if it may contain information about

- A case where a body may eventually be found
- A missing person report
- A solved case about a missing person

Respond with "true" or "false" depending on if the video should be downloaded (true) or not (false).
`

	// stringify struct
	json, err := json.Marshal(article)
	if err != nil {
		log.Err(err).Msg("failed to marshal article")
		return false, err
	}

	// call local ollama API at http://localhost:11434/v1/chat
	response, err := CallOllama("llama3.1", systemPrompt, string(json))
	if err != nil {
		log.Err(err).Msg("failed to call ollama")
		return false, err
	}
	// if theres more than 5 letters in the response, its probably not a valid response
	if len(response) > len("false") {
		return false, fmt.Errorf("response is too long, not a valid response, got %s", response)
	}
	// check if response is true or false
	return response == "true", nil

}

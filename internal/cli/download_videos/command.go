package download_videos

import (
	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	use   = "download-videos"
	short = "Download videos that have been approved by AI check"
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

	// Find all articles where AI suggests downloading, regardless of provider
	filter := bson.M{
		"aiSuggestsDownloadingVideo": true,
	}

	articles, err := db.Models.FindArticlesByFilter(ctx, filter)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find articles")
	}

	if len(articles) == 0 {
		log.Info().Msg("no videos to download")
		return
	}

	log.Info().Int("count", len(articles)).Msg("found videos to download")

	for _, article := range articles {
		log.Info().Str("url", article.Url).Str("title", article.Title).Msg("downloading video")

		switch article.Provider {
		case "cnn":
			err := downloadCnnVideo(article)
			if err != nil {
				log.Err(err).Str("url", article.Url).Msg("failed to download video, skipping")
				continue
			}
			log.Info().Str("url", article.Url).Msg("successfully downloaded video")
		default:
			log.Warn().Str("provider", article.Provider).Msg("unsupported provider for video download")
		}
	}
}

// downloadCnnVideo is a placeholder function that would actually download the video
func downloadCnnVideo(article *db.Article) error {
	// This is just a dummy implementation
	log.Info().Str("url", article.Url).Msg("simulating video download")

	return nil
}

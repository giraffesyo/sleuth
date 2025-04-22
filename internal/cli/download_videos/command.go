package download_videos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	use   = "download-videos"
	short = "Download videos that have been approved by AI check"
	// Directory to store downloaded videos
	downloadDir = "./downloads"
)

var Cmd = &cobra.Command{
	Use:   use,
	Short: short,
	Run:   run,
}

func init() {
	// Create downloads directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("failed to create downloads directory")
	}
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

		err := determineVideoUrl(ctx, article)
		if err != nil {
			log.Err(err).Str("url", article.Url).Msg("failed to determine video URL, skipping")
			continue
		}
		// Check if the video has already been downloaded
		if _, err := os.Stat(article.VideoPath); err == nil {
			log.Info().Str("url", article.Url).Str("path", article.VideoPath).Msg("video already downloaded, skipping")
			continue
		}
		log.Info().Str("url", article.Url).Str("title", article.Title).Msg("downloading video")

		switch article.Provider {
		case "cnn":
			videoPath, err := downloadVideo(article, article.VideoUrl)
			if err != nil {
				log.Err(err).Str("url", article.Url).Msg("failed to download video, skipping")
				continue
			}

			log.Info().Str("url", article.Url).Str("path", videoPath).Msg("successfully downloaded video and updated database")
		default:
			log.Warn().Str("provider", article.Provider).Msg("unsupported provider for video download")
		}
	}
}

// getVideoFilePath returns the path where the video file for an article should be stored
func getVideoFilePath(article *db.Article, extension string) string {
	// Use the article's database ID as the filename
	filename := article.Id.Hex() + extension
	return filepath.Join(downloadDir, filename)
}

// CnnVideoResponse represents the JSON response from CNN's video API
type CnnVideoResponse struct {
	Files []struct {
		FileUri string `json:"fileUri"`
	} `json:"files"`
	Headline string `json:"headline"`
	Id       string `json:"id"`
}

// will update the article with the video URL and path
// if the video URL is not already set
func determineVideoUrl(ctx context.Context, article *db.Article) error {
	if article.VideoUrl != "" && article.VideoPath != "" {
		return nil
	}
	var videoUrl string
	var err error
	switch article.Provider {
	case "cnn":
		videoUrl, err = determineCnnVideoUrl(ctx, article)
	default:
		return fmt.Errorf("unsupported provider: %s", article.Provider)
	}
	if err != nil {
		return fmt.Errorf("failed to determine video URL: %w", err)
	}
	// extract extension from the URL
	fileExt := filepath.Ext(videoUrl)
	if fileExt == "" {
		fileExt = ".mp4" // Default extension
	}
	videoPath := getVideoFilePath(article, fileExt)
	// update the article with the video URL
	update := bson.M{
		"videoUrl":  videoUrl,
		"videoPath": videoPath,
	}
	if err := db.Models.UpdateArticle(ctx, article.Id, update); err != nil {
		return fmt.Errorf("failed to update article with video URL: %w", err)
	}
	// update local article object
	article.VideoUrl = videoUrl
	article.VideoPath = videoPath
	// Return the video URL
	return nil

}

// determineVideoUrl extracts the direct video URL for a CNN article
// This handles the browser automation and API calls to get the URL
func determineCnnVideoUrl(ctx context.Context, article *db.Article) (string, error) {
	// Create a new ChromeDP context for browser automation
	chromectx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// Set a timeout
	chromectx, cancel = context.WithTimeout(chromectx, 60*time.Second)
	defer cancel()

	log.Info().Str("url", article.Url).Msg("navigating to article URL with ChromeDP")

	var videoUri string
	// Navigate to the page and extract the video URI using JavaScript
	err := chromedp.Run(chromectx,
		chromedp.Navigate(article.Url),
		// Wait for the video element to be present
		chromedp.WaitVisible(`div[data-video-id]`, chromedp.ByQuery),
		// Execute JavaScript to get the URI
		chromedp.Evaluate(`document.querySelector("div[data-video-id]").dataset.uri`, &videoUri),
	)
	if err != nil {
		return "", fmt.Errorf("failed to extract video URI using ChromeDP: %w", err)
	}

	if videoUri == "" {
		return "", fmt.Errorf("couldn't find video URI in the article page")
	}
	log.Info().Str("videoUri", videoUri).Msg("found video URI")

	// Construct the API URL
	apiURL := fmt.Sprintf("https://fave.api.cnn.io/v1/video?id=111111&stellarUri=%s", videoUri)
	log.Info().Str("apiURL", apiURL).Msg("fetching video metadata")

	// Fetch the video metadata
	videoResp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch video metadata: %w", err)
	}
	defer videoResp.Body.Close()

	// Parse the JSON response
	var videoData CnnVideoResponse
	if err := json.NewDecoder(videoResp.Body).Decode(&videoData); err != nil {
		return "", fmt.Errorf("failed to decode video metadata: %w", err)
	}

	// Check if we have a file URL
	if len(videoData.Files) == 0 {
		return "", fmt.Errorf("no video files found in the metadata")
	}

	// Get the direct MP4 URL
	videoURL := videoData.Files[0].FileUri
	log.Info().Str("videoURL", videoURL).Msg("found video URL")

	return videoURL, nil
}

// downloadVideo downloads a video from the provided URL and saves it to the filesystem
// Returns the path to the downloaded video file
func downloadVideo(article *db.Article, videoURL string) (string, error) {
	// Extract file extension from the URL
	fileExt := filepath.Ext(videoURL)
	if fileExt == "" {
		fileExt = ".mp4" // Default extension
	}

	// Get the complete file path
	fullPath := getVideoFilePath(article, fileExt)
	log.Info().Str("path", fullPath).Msg("downloading video to file")

	// Download the video file
	videoFileResp, err := http.Get(videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download video file: %w", err)
	}
	defer videoFileResp.Body.Close()

	// Create the output file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy the video data to the file
	_, err = io.Copy(outFile, videoFileResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save video file: %w", err)
	}

	log.Info().Str("path", fullPath).Msg("video downloaded successfully")
	return fullPath, nil
}

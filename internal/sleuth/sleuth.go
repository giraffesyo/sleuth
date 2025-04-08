package sleuth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/giraffesyo/sleuth/internal/sleuth/providers"
	"github.com/giraffesyo/sleuth/internal/sleuth/providers/cnn"
	"github.com/giraffesyo/sleuth/internal/sleuth/providers/fox"
	"github.com/giraffesyo/sleuth/internal/sleuth/videos"
	"github.com/rs/zerolog/log"
)

type sleuth struct {
	enabledProviders []string
	query            string
}

type sleuthOption func(*sleuth)

func WithProvider(p ...string) sleuthOption {
	return func(s *sleuth) {
		s.enabledProviders = p
	}
}

func WithSearchQuery(query string) sleuthOption {
	return func(s *sleuth) {
		s.query = query
	}
}

func NewSleuth(options ...sleuthOption) *sleuth {
	s := &sleuth{}
	for _, o := range options {
		o(s)
	}
	return s
}

var dbClient *db.Mongo

func writeVideosToFile(provider string, videos []videos.Video) error {
	jsonData, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal videos: %w", err)
	}
	filename := fmt.Sprintf("%s.json", provider)
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write videos to file: %w", err)
	}
	return nil
}

func (s *sleuth) Run() error {
	ctx := context.Background()
	if s.query == "" {
		return ErrEmptySearchQuery
	}

	// initialize the database client
	dbClient = db.NewMongo()
	if err := dbClient.ConnectDatabase("mongodb://localhost:27017"); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().Str("query", s.query).Msg("searching for news articles")
	for _, p := range s.enabledProviders {
		var provider providers.Provider
		switch p {
		case cnn.ProviderCNN:
			log.Info().Msg("CNN is enabled")
			provider = cnn.NewCNNProvider(ctx)
		case fox.ProviderFoxNews:
			log.Info().Msg("Fox News is enabled")
			provider = fox.NewFoxProvider()
		default:
			log.Warn().Str("provider", p).Msg("unknown provider, skipping")
			continue
		}
		videos, err := provider.Search(s.query)
		if err != nil {
			return fmt.Errorf("failed to search CNN: %w", err)
		}
		log.Info().Str("provider", provider.ProviderName()).Int("count", len(videos)).Msg("search results")
		// write videos to file
		err = writeVideosToFile(provider.ProviderName(), videos)
		if err != nil {
			return fmt.Errorf("failed to write videos to file: %w", err)
		}
	}
	return nil
}

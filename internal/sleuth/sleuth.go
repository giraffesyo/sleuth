package sleuth

import (
	"context"

	"github.com/giraffesyo/sleuth/internal/sleuth/providers"
	"github.com/rs/zerolog/log"
)

type sleuth struct {
	enabledProviders []provider
	query            string
}

type sleuthOption func(*sleuth)

type provider string

const ProviderCNN = provider("cnn")
const ProviderFoxNews = provider("foxnews")

func WithProvider(p provider) sleuthOption {
	return func(s *sleuth) {
		s.enabledProviders = append(s.enabledProviders, p)
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

func (s *sleuth) Run() error {
	ctx := context.Background()
	if s.query == "" {
		return ErrEmptySearchQuery
	}
	log.Info().Str("query", s.query).Msg("searching for news articles")
	for _, p := range s.enabledProviders {
		switch p {
		case ProviderCNN:
			log.Info().Msg("CNN is enabled")
			providers.NewCNNProvider(ctx).Search(s.query)
		case ProviderFoxNews:
			log.Info().Msg("Fox News is enabled")
			log.Warn().Msg("Fox News provider is not implemented")
		}
	}
	return nil
}

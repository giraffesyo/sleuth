package sleuth

import (
	"github.com/rs/zerolog/log"
)

type sleuth struct {
	enabledProviders []provider
	searchTerm       string
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

func WithSearchTerm(term string) sleuthOption {
	return func(s *sleuth) {
		s.searchTerm = term
	}
}

func NewSleuth(options ...sleuthOption) *sleuth {
	s := &sleuth{}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *sleuth) Run() {
	for _, p := range s.enabledProviders {
		switch p {
		case ProviderCNN:
			log.Info().Msg("CNN is enabled")
		case ProviderFoxNews:
			log.Info().Msg("Fox News is enabled")
		}
	}
}

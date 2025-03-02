package fox

import (
	"github.com/giraffesyo/sleuth/internal/sleuth/providers"
	"github.com/giraffesyo/sleuth/internal/sleuth/videos"
)

const ProviderFoxNews = "foxnews"

type foxProviderOption func(*foxProvider)

type foxProvider struct {
	searchUrl      string
	withPagination bool
}

func WithCustomSearchUrl(url string) foxProviderOption {
	return func(p *foxProvider) {
		p.searchUrl = url
	}
}

func WithoutPagination() foxProviderOption {
	return func(p *foxProvider) {
		p.withPagination = false
	}
}

func NewFoxProvider(providerOptions ...foxProviderOption) *foxProvider {
	p := &foxProvider{
		searchUrl:      "https://www.foxnews.com/search?q=",
		withPagination: true,
	}
	for _, o := range providerOptions {
		o(p)
	}
	return p
}

func (p *foxProvider) ProviderName() string {
	return ProviderFoxNews
}

func (p *foxProvider) Search(query string) ([]videos.Video, error) {
	return nil, nil
}

// ensure foxProvider implements the Provider interface
var _ providers.Provider = &foxProvider{}

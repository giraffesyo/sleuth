package providers

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/giraffesyo/sleuth/internal/sleuth/videos"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/rs/zerolog/log"
)

const ProviderCNN = "cnn"

type Provider interface {
	Search(query string) ([]videos.Video, error)
	ProviderName() string
}

type providerOption func(*cnnProvider)

type cnnProvider struct {
	context        context.Context
	allowedDomains []string
	searchUrl      string
}

// Used for testing purposes, to allow the test to serve cnn from custom domain.
func WithCustomSearchUrl(url string) providerOption {
	return func(p *cnnProvider) {
		p.searchUrl = url
		// parse the domain from the URL
		domain := strings.TrimPrefix(url, "https://")
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.Split(domain, "/")[0]

		p.allowedDomains = append(p.allowedDomains, domain)
	}
}

func NewCNNProvider(ctx context.Context, providerOptions ...providerOption) *cnnProvider {
	p := &cnnProvider{
		context:        ctx,
		allowedDomains: []string{"cnn.com", "www.cnn.com"},
		searchUrl:      "https://www.cnn.com/search?types=video&q=",
	}
	for _, o := range providerOptions {
		o(p)
	}
	return p
}

func (p *cnnProvider) ProviderName() string {
	return ProviderCNN
}

func (p *cnnProvider) Search(query string) ([]videos.Video, error) {
	c := colly.NewCollector(
		colly.AllowedDomains(p.allowedDomains...),
		colly.Debugger(&debug.LogDebugger{}),
	)
	// url escape the query
	escapedquery := url.QueryEscape(query)
	c.OnRequest(func(r *colly.Request) {
		// simply log each request
		log.Info().Str("url", r.URL.String()).Msg("visiting")
	})

	results := []videos.Video{}
	// Callback for when an HTML element with the video result appears.
	c.OnHTML(`div[data-uri^="/_components/card/instances/search-"]`, func(e *colly.HTMLElement) {
		// Extract the video link from the <a> element with the appropriate class.
		link := e.ChildAttr("a.container__link--type-Video", "href")
		if link == "" {
			return
		}
		// Ensure the URL is absolute.
		if strings.HasPrefix(link, "/") {
			link = "https://www.cnn.com" + link
		}

		// Use "/video/" as the heuristic to confirm the result is a video.
		if !strings.Contains(link, "/video/") {
			return
		}

		// Get the title, date, and description.
		title := e.ChildText("span.container__headline-text")
		date := e.ChildText("div.container__date")
		description := e.ChildText("div.container__description")

		video := videos.Video{
			URL:         link,
			Title:       title,
			Date:        date,
			Description: description,
			Provider:    ProviderCNN,
		}
		results = append(results, video)
	})

	err := c.Visit(fmt.Sprintf("%s%s", p.searchUrl, escapedquery))
	if err != nil {
		return nil, err
	}

	c.Wait()

	return results, nil
}

// ensure that CNN implements the Provider interface
var _ Provider = &cnnProvider{}

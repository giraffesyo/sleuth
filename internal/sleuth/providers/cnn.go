package providers

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/giraffesyo/sleuth/internal/sleuth/videos"
	"github.com/rs/zerolog/log"
)

const ProviderCNN = "cnn"

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
	// Create a new chromedp context from the provider's context.
	ctx, cancel := chromedp.NewContext(p.context)
	defer cancel()

	// Set a timeout for the chromedp tasks.
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Build the search URL.
	escapedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s%s", p.searchUrl, escapedQuery)
	log.Info().Str("url", searchURL).Msg("Navigating to search URL with chromedp")

	var renderedHTML string
	// Run chromedp tasks:
	// 1. Navigate to the search URL.
	// 2. Wait until at least one search result card is visible.
	// 3. Extract the outer HTML of the page.
	tasks := chromedp.Tasks{
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`div[data-uri^="/_components/card/instances/search-"]`, chromedp.ByQuery),
		chromedp.OuterHTML("html", &renderedHTML, chromedp.ByQuery),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		log.Error().Err(err).Msg("chromedp run failed")
		return nil, err
	}

	// Parse the rendered HTML using goquery.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(renderedHTML))
	if err != nil {
		return nil, err
	}

	results := []videos.Video{}
	// Iterate over each search result card.
	doc.Find(`div[data-uri^="/_components/card/instances/search-"]`).Each(func(i int, s *goquery.Selection) {
		link, exists := s.Find("a.container__link--type-Video").Attr("href")
		if !exists || link == "" {
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

		title := s.Find("span.container__headline-text").Text()
		date := s.Find("div.container__date").Text()
		description := s.Find("div.container__description").Text()

		video := videos.Video{
			URL:         strings.TrimSpace(link),
			Title:       strings.TrimSpace(title),
			Date:        strings.TrimSpace(date),
			Description: strings.TrimSpace(description),
			Provider:    ProviderCNN,
		}
		results = append(results, video)
	})

	return results, nil
}

// ensure that CNN implements the Provider interface
var _ Provider = &cnnProvider{}

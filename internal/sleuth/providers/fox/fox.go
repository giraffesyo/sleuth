package fox

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/giraffesyo/sleuth/internal/sleuth/providers"
	"github.com/giraffesyo/sleuth/internal/sleuth/videos"
	"github.com/rs/zerolog/log"
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
		searchUrl:      "https://www.foxnews.com/search-results/search#q=",
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
	// Create a chromedp context.
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set an overall timeout.
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	escapedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s%s", p.searchUrl, escapedQuery)
	log.Info().Str("url", searchURL).Msg("Navigating to Fox News search URL with chromedp")

	// Navigate to the search URL and wait until at least one article is visible.
	if err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`article.article`, chromedp.ByQuery),
	); err != nil {
		return nil, err
	}

	// Use a map to deduplicate articles by URL.
	seen := make(map[string]struct{})
	var allResults []videos.Video

	// extractArticles parses the provided HTML and appends new articles.
	extractArticles := func(html string) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			log.Error().Err(err).Msg("failed to create goquery document")
			return
		}
		doc.Find("article.article").Each(func(i int, s *goquery.Selection) {
			// Get the article URL from the <a> inside the "m" container.
			a := s.Find("div.m a")
			link, exists := a.Attr("href")
			if !exists || link == "" {
				return
			}
			link = strings.TrimSpace(link)
			// Avoid duplicates.
			if _, found := seen[link]; found {
				return
			}
			seen[link] = struct{}{}

			// Extract title from the <h2 class="title"><a> element.
			title := strings.TrimSpace(s.Find("h2.title a").Text())
			// Extract date from the <span class="time"> inside the meta section.
			date := strings.TrimSpace(s.Find("header.info-header div.meta span.time").Text())
			// Extract description from the <p class="dek">.
			description := strings.TrimSpace(s.Find("div.content p.dek").Text())

			video := videos.Video{
				URL:         link,
				Title:       title,
				Date:        date,
				Description: description,
				Provider:    ProviderFoxNews,
			}
			log.Debug().Str("title", video.Title).Str("provider", p.ProviderName()).Str("date", video.Date).Str("url", video.URL).Msg("Found video")
			allResults = append(allResults, video)
		})
	}

	// Get the initial rendered HTML.
	var renderedHTML string
	if err := chromedp.Run(ctx,
		chromedp.OuterHTML("html", &renderedHTML, chromedp.ByQuery),
	); err != nil {
		return nil, err
	}
	extractArticles(renderedHTML)

	// Loop to click "Load More" until the button is no longer present.
	if p.withPagination {
		for {
			var loadMoreExists bool
			checkJS := `document.querySelector('div.button.load-more a') !== null`
			if err := chromedp.Run(ctx, chromedp.Evaluate(checkJS, &loadMoreExists)); err != nil {
				log.Error().Err(err).Msg("failed to evaluate load more existence")
				break
			}
			if !loadMoreExists {
				break
			}
			log.Debug().Str("provider", p.ProviderName()).Msg("Going to next page of results")
			// Click the "Load More" button.
			if err := chromedp.Run(ctx,
				chromedp.Click(`div.button.load-more a`, chromedp.ByQuery),
				chromedp.Sleep(2*time.Second), // wait for the new articles to load
			); err != nil {
				log.Error().Err(err).Msg("failed to click load more")
				break
			}

			// Get the updated HTML.
			if err := chromedp.Run(ctx,
				chromedp.OuterHTML("html", &renderedHTML, chromedp.ByQuery),
			); err != nil {
				log.Error().Err(err).Msg("failed to get updated HTML")
				break
			}
			extractArticles(renderedHTML)
		}
	}

	return allResults, nil
}

// ensure foxProvider implements the Provider interface
var _ providers.Provider = &foxProvider{}

package cnn

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/giraffesyo/sleuth/internal/db"
	"github.com/giraffesyo/sleuth/internal/sleuth/providers"
	"github.com/rs/zerolog/log"
)

const ProviderCNN = "cnn"

type providerOption func(*cnnProvider)

type cnnProvider struct {
	context        context.Context
	searchUrl      string
	withPagination bool
}

// Used for testing purposes, to allow the test to serve cnn from custom domain.
func WithCustomSearchUrl(url string) providerOption {
	return func(p *cnnProvider) {
		p.searchUrl = url
	}
}

func WithoutPagination() providerOption {
	return func(p *cnnProvider) {
		p.withPagination = false
	}
}

func NewCNNProvider(ctx context.Context, providerOptions ...providerOption) *cnnProvider {
	p := &cnnProvider{
		context:        ctx,
		searchUrl:      "https://www.cnn.com/search?types=video&q=",
		withPagination: true,
	}
	for _, o := range providerOptions {
		o(p)
	}
	return p
}

func (p *cnnProvider) ProviderName() string {
	return ProviderCNN
}

func (p *cnnProvider) Search(query string) ([]db.Article, error) {
	// Create a chromedp context using the provider's context.
	ctx, cancel := chromedp.NewContext(p.context)
	defer cancel()

	// Set an overall timeout.
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	escapedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s%s", p.searchUrl, escapedQuery)
	log.Info().Str("url", searchURL).Msg("Navigating to search URL with chromedp")

	// Navigate to the search page and wait for the results to load.
	if err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.WaitVisible(`div[data-uri^="/_components/card/instances/search-"]`, chromedp.ByQuery),
	); err != nil {
		return nil, err
	}

	var allResults []db.Article

	// Loop to process each page.
	for {
		// Extract the full rendered HTML.
		var renderedHTML string
		if err := chromedp.Run(ctx,
			chromedp.OuterHTML("html", &renderedHTML, chromedp.ByQuery),
		); err != nil {
			return nil, err
		}

		// Parse the HTML using goquery.
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(renderedHTML))
		if err != nil {
			return nil, err
		}

		// Extract video details from each card.
		doc.Find(`div[data-uri^="/_components/card/instances/search-"]`).Each(func(i int, s *goquery.Selection) {
			link, exists := s.Find("a.container__link--type-Video").Attr("href")
			if !exists || link == "" {
				return
			}
			if strings.HasPrefix(link, "/") {
				link = "https://www.cnn.com" + link
			}

			title := strings.TrimSpace(s.Find("span.container__headline-text").Text())
			date := strings.TrimSpace(s.Find("div.container__date").Text())
			description := strings.TrimSpace(s.Find("div.container__description").Text())

			article := db.Article{
				Url:         link,
				Title:       title,
				Date:        date,
				Description: description,
				Provider:    p.ProviderName(),
			}
			err := db.Models.CreateArticle(p.context, &article)
			if err != nil {
				log.Error().Err(err).Msg("Failed to save video to database")
				return
			}
			log.Debug().Str("title", article.Title).Str("provider", p.ProviderName()).Str("date", article.Date).Str("url", article.Url).Msg("Found video")
			allResults = append(allResults, article)
		})

		// Check if a "Next" button is available by verifying if the element with active classes exists.
		var hasNext bool
		evaluateJS := `document.querySelector('div.pagination-arrow.pagination-arrow-right.search__pagination-link.text-active') !== null`
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(evaluateJS, &hasNext),
		); err != nil {
			return nil, err
		}

		// If no next page or pagination is disabled, break out of the loop.
		if !hasNext || !p.withPagination {
			break
		}
		log.Debug().Msg("Going to next page of results")
		// Click the "Next" button.
		if err := chromedp.Run(ctx,
			chromedp.Click(`div.pagination-arrow.pagination-arrow-right.search__pagination-link.text-active`, chromedp.ByQuery),
			// Give the page time to load the new results.
			chromedp.Sleep(2*time.Second),
			chromedp.WaitVisible(`div[data-uri^="/_components/card/instances/search-"]`, chromedp.ByQuery),
		); err != nil {
			return nil, err
		}
	}

	return allResults, nil
}

// ensure that CNN implements the Provider interface
var _ providers.Provider = &cnnProvider{}

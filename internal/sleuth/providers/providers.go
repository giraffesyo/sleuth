package providers

import "github.com/giraffesyo/sleuth/internal/sleuth/videos"

type Provider interface {
	Search(query string) ([]videos.Video, error)
	ProviderName() string
}

package providers

import (
	"github.com/giraffesyo/sleuth/internal/db"
)

type Provider interface {
	Search(query string) ([]db.Article, error)
	ProviderName() string
}

package providers

import "context"

type Provider interface {
	Search(term string) error
}

type cnnProvider struct {
	context context.Context
}

func NewCNNProvider(ctx context.Context) *cnnProvider {

	return &cnnProvider{}
}

func (c *cnnProvider) Search(term string) error {

	return nil
}

// ensure that CNN implements the Provider interface
var _ Provider = &cnnProvider{}

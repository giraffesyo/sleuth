package providers

type Provider interface {
	Search(term string) error
}

type CNN struct {
}

func (c *CNN) Search(term string) error {
	return nil
}

// ensure that CNN implements the Provider interface
var _ Provider = &CNN{}

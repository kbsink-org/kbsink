package core

import "context"

// Driver fetches article page source by URL.
type Driver interface {
	Fetch(ctx context.Context, url string) (*FetchResult, error)
}

// Parser transforms raw HTML into metadata, markdown, and asset references.
type Parser interface {
	Parse(ctx context.Context, fetched *FetchResult, outputDir string) (*ArticleResult, error)
}

// Storage persists article markdown and media assets.
type Storage interface {
	Save(ctx context.Context, article *ArticleResult) error
}

// Plugin is a named pair of Parser and Driver for registry-based wiring.
type Plugin interface {
	Name() string
	NewComponents(client HTTPClient) (Parser, Driver, error)
}

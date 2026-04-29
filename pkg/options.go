package kbsink

import (
	"net/http"

	drv "github.com/kbsink-org/kbsink/pkg/driver"
	prs "github.com/kbsink-org/kbsink/pkg/parser"
	stg "github.com/kbsink-org/kbsink/pkg/storage"
)

type converterConfig struct {
	driver drv.Driver
	parser prs.Parser
	store  stg.Storage
	client *http.Client
}

// Option configures a Converter.
type Option func(*converterConfig)

func WithDriver(d drv.Driver) Option {
	return func(c *converterConfig) { c.driver = d }
}

func WithParser(p prs.Parser) Option {
	return func(c *converterConfig) { c.parser = p }
}

func WithStorage(s stg.Storage) Option {
	return func(c *converterConfig) { c.store = s }
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *converterConfig) { c.client = client }
}

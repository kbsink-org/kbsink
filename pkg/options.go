package kbsink

import (
	"net/http"

	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/logger"
)

type converterConfig struct {
	driver      core.Driver
	parser      core.Parser
	store       core.Storage
	client      *http.Client
	log         logger.Logger
	minLevel    logger.Level
	minLevelSet bool
}

// Option configures a Converter.
type Option func(*converterConfig)

func WithDriver(d core.Driver) Option {
	return func(c *converterConfig) { c.driver = d }
}

func WithParser(p core.Parser) Option {
	return func(c *converterConfig) { c.parser = p }
}

func WithStorage(s core.Storage) Option {
	return func(c *converterConfig) { c.store = s }
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *converterConfig) { c.client = client }
}

// WithLogger injects pipeline logging for converter, driver, parser, and storage.
// Nil uses [logger.Default] (nop unless [logger.SetDefault] was called).
func WithLogger(l logger.Logger) Option {
	return func(c *converterConfig) { c.log = l }
}

// WithMinLevel drops log lines below min (e.g. WithMinLevel(logger.LevelWarn) hides debug/info).
func WithMinLevel(min logger.Level) Option {
	return func(c *converterConfig) {
		c.minLevel = min
		c.minLevelSet = true
	}
}

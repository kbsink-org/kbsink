package kbsink

import (
	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/logger"
)

type converterConfig struct {
	driver      core.Driver
	parser      core.Parser
	store       core.Storage
	client      core.HTTPClient
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

func WithHTTPClient(client core.HTTPClient) Option {
	return func(c *converterConfig) { c.client = client }
}

// WithLogger enables pipeline logging on the converter and on built-in HTMLDriver/LocalStorage.
// Nil or omitted means no logs. Injected driver/parser/storage are not wrapped.
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

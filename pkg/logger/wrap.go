package logger

import (
	"context"

	"github.com/kbsink-org/kbsink/pkg/core"
)

// WrapDriver logs fetch lifecycle around inner.
func WrapDriver(inner core.Driver, log Logger, component string) core.Driver {
	if inner == nil {
		return nil
	}
	log = Resolve(log)
	if component == "" {
		component = "driver"
	}
	return &loggingDriver{inner: inner, log: log, component: component}
}

// WrapParser logs parse lifecycle around inner.
func WrapParser(inner core.Parser, log Logger, component string) core.Parser {
	if inner == nil {
		return nil
	}
	log = Resolve(log)
	if component == "" {
		component = "parser"
	}
	return &loggingParser{inner: inner, log: log, component: component}
}

// WrapStorage logs save lifecycle around inner.
func WrapStorage(inner core.Storage, log Logger, component string) core.Storage {
	if inner == nil {
		return nil
	}
	log = Resolve(log)
	if component == "" {
		component = "storage"
	}
	return &loggingStorage{inner: inner, log: log, component: component}
}

type loggingDriver struct {
	inner     core.Driver
	log       Logger
	component string
}

func (d *loggingDriver) Fetch(ctx context.Context, url string) (*core.FetchResult, error) {
	d.log.Debug(d.component+": fetch start", "url", url)
	res, err := d.inner.Fetch(ctx, url)
	if err != nil {
		d.log.Error(d.component+": fetch failed", "url", url, "err", err)
		return nil, err
	}
	htmlLen := 0
	finalURL := url
	if res != nil {
		htmlLen = len(res.HTML)
		if res.URL != "" {
			finalURL = res.URL
		}
	}
	d.log.Info(d.component+": fetch done", "url", finalURL, "htmlLen", htmlLen)
	return res, nil
}

type loggingParser struct {
	inner     core.Parser
	log       Logger
	component string
}

func (p *loggingParser) Parse(ctx context.Context, fetched *core.FetchResult, outputDir string) (*core.ArticleResult, error) {
	srcURL := ""
	if fetched != nil {
		srcURL = fetched.URL
	}
	p.log.Debug(p.component+": parse start", "url", srcURL, "outputDir", outputDir)
	res, err := p.inner.Parse(ctx, fetched, outputDir)
	if err != nil {
		p.log.Error(p.component+": parse failed", "url", srcURL, "err", err)
		return nil, err
	}
	assetCount := 0
	title := ""
	if res != nil {
		title = res.Title
		assetCount = len(res.Assets)
	}
	p.log.Info(p.component+": parse done", "url", srcURL, "title", title, "assets", assetCount)
	return res, nil
}

type loggingStorage struct {
	inner     core.Storage
	log       Logger
	component string
}

func (s *loggingStorage) Save(ctx context.Context, article *core.ArticleResult) error {
	dir, title := "", ""
	assetCount := 0
	if article != nil {
		dir = article.OutputDir
		title = article.Title
		assetCount = len(article.Assets)
	}
	s.log.Debug(s.component+": save start", "outputDir", dir, "title", title, "assets", assetCount)
	err := s.inner.Save(ctx, article)
	if err != nil {
		s.log.Error(s.component+": save failed", "outputDir", dir, "err", err)
		return err
	}
	s.log.Info(s.component+": save done", "outputDir", dir, "markdownPath", articleMarkdownPath(article))
	return nil
}

func articleMarkdownPath(article *core.ArticleResult) string {
	if article == nil {
		return ""
	}
	return article.MarkdownPath
}

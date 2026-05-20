package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/logger"
)

const defaultOutputRoot = "output"

// LocalStorage saves markdown and images into local filesystem.
type LocalStorage struct {
	root string
	log  logger.Logger
}

func NewLocalStorage(root string, log logger.Logger) *LocalStorage {
	if root == "" {
		root = defaultOutputRoot
	}
	return &LocalStorage{root: root, log: log}
}

func (s *LocalStorage) Save(_ context.Context, article *core.ArticleResult) error {
	if article == nil {
		return fmt.Errorf("article is nil")
	}
	if s.log != nil {
		s.log.Debug("local storage: save start", "outputDir", article.OutputDir, "title", article.Title)
	}

	baseDir := filepath.FromSlash(article.OutputDir)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		if s.log != nil {
			s.log.Error("local storage: mkdir failed", "dir", baseDir, "err", err)
		}
		return err
	}
	assets := article.Assets
	if len(assets) == 0 && len(article.Images) > 0 {
		assets = make([]core.Asset, 0, len(article.Images))
		for _, img := range article.Images {
			assets = append(assets, core.Asset{
				Type:         core.AssetTypeImage,
				SourceURL:    img.SourceURL,
				RelativePath: img.RelativePath,
				FileName:     img.FileName,
				ContentType:  img.ContentType,
				Data:         img.Data,
			})
		}
	}

	for _, asset := range assets {
		target := filepath.Join(baseDir, filepath.FromSlash(asset.RelativePath))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			if s.log != nil {
				s.log.Error("local storage: asset mkdir failed", "path", target, "err", err)
			}
			return err
		}
		if err := os.WriteFile(target, asset.Data, 0o644); err != nil {
			if s.log != nil {
				s.log.Error("local storage: write asset failed", "path", target, "err", err)
			}
			return err
		}
		if s.log != nil {
			s.log.Debug("local storage: wrote asset", "path", target, "bytes", len(asset.Data))
		}
	}
	mdPath := filepath.FromSlash(article.MarkdownPath)
	if err := os.WriteFile(mdPath, []byte(article.Markdown), 0o644); err != nil {
		if s.log != nil {
			s.log.Error("local storage: write markdown failed", "path", mdPath, "err", err)
		}
		return err
	}
	if s.log != nil {
		s.log.Info("local storage: save done", "markdownPath", mdPath, "assets", len(assets))
	}
	return nil
}

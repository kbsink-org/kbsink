package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd/core"
)

const defaultOutputRoot = "output"

// LocalStorage saves markdown and images into local filesystem.
type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) *LocalStorage {
	if root == "" {
		root = defaultOutputRoot
	}
	return &LocalStorage{root: root}
}

func (s *LocalStorage) Save(_ context.Context, article *core.ArticleResult) error {
	if article == nil {
		return fmt.Errorf("article is nil")
	}
	baseDir := filepath.FromSlash(article.OutputDir)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
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
			return err
		}
		if err := os.WriteFile(target, asset.Data, 0o644); err != nil {
			return err
		}
	}
	mdPath := filepath.FromSlash(article.MarkdownPath)
	if err := os.WriteFile(mdPath, []byte(article.Markdown), 0o644); err != nil {
		return err
	}
	return nil
}

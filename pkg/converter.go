package kbsink

import (
	"context"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kbsink-org/kbsink/pkg/core"
	drv "github.com/kbsink-org/kbsink/pkg/driver"
	prs "github.com/kbsink-org/kbsink/pkg/parser"
	stg "github.com/kbsink-org/kbsink/pkg/storage"
)

const (
	defaultOutputRoot = "output"
)

var (
	unsafeFileChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	spaceChars      = regexp.MustCompile(`\s+`)
)

// Converter orchestrates driver -> parser -> image fetch -> storage.
type Converter struct {
	driver drv.Driver
	parser prs.Parser
	store  stg.Storage
	client *http.Client
}

// NewConverter creates a converter with sensible defaults.
//
// By default it uses: HTMLDriver + WechatParser + LocalStorage + http.DefaultClient.
func NewConverter(opts ...Option) *Converter {
	cfg := &converterConfig{
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	if cfg.driver == nil {
		cfg.driver = drv.NewHTMLDriver(cfg.client)
	}
	if cfg.parser == nil {
		cfg.parser = prs.NewWechatParser()
	}
	if cfg.store == nil {
		cfg.store = stg.NewLocalStorage(defaultOutputRoot)
	}

	return &Converter{
		driver: cfg.driver,
		parser: cfg.parser,
		store:  cfg.store,
		client: cfg.client,
	}
}

// Convert fetches one WeChat article, converts markdown, downloads images, and saves output.
func (c *Converter) Convert(ctx context.Context, articleURL string, opts core.ConvertOptions) (*core.ArticleResult, error) {
	if articleURL == "" {
		return nil, fmt.Errorf("article url is required")
	}

	outputRoot := opts.OutputRoot
	if outputRoot == "" {
		outputRoot = defaultOutputRoot
	}
	videoMode := opts.VideoMode
	if videoMode == "" {
		videoMode = core.VideoModeLink
	}

	raw, err := c.driver.Fetch(ctx, articleURL)
	if err != nil {
		return nil, fmt.Errorf("fetch article: %w", err)
	}

	outDir := outputRoot
	parsed, err := c.parser.Parse(ctx, raw, outDir)
	if err != nil {
		return nil, fmt.Errorf("parse article: %w", err)
	}

	parsed.SafeTitle = sanitizeFileName(parsed.Title)
	if parsed.SafeTitle == "" {
		parsed.SafeTitle = "untitled_article"
	}
	parsed.OutputDir = filepath.ToSlash(path.Join(outputRoot, parsed.SafeTitle))
	parsed.MarkdownPath = filepath.ToSlash(path.Join(parsed.OutputDir, parsed.SafeTitle+".md"))

	assets := parsed.Assets
	if len(assets) == 0 && len(parsed.Images) > 0 {
		assets = make([]core.Asset, 0, len(parsed.Images))
		for _, img := range parsed.Images {
			assets = append(assets, core.Asset{
				Type:      core.AssetTypeImage,
				SourceURL: img.SourceURL,
			})
		}
	}
	imageIdx := 0
	videoIdx := 0
	for i := range assets {
		data, contentType, ext, dlErr := c.downloadAsset(ctx, assets[i].SourceURL)
		if dlErr != nil {
			return nil, fmt.Errorf("download asset %q: %w", assets[i].SourceURL, dlErr)
		}
		assetType := assets[i].Type
		if assetType == "" {
			assetType = inferAssetType(contentType)
		}
		if assetType == "" {
			assetType = core.AssetTypeImage
		}

		var idx int
		var fileName string
		var relativePath string
		switch assetType {
		case core.AssetTypeVideo:
			videoIdx++
			idx = videoIdx
			fileName = fmt.Sprintf("video_%03d%s", idx, ext)
			relativePath = filepath.ToSlash(path.Join("videos", fileName))
		default:
			imageIdx++
			idx = imageIdx
			fileName = fmt.Sprintf("img_%03d%s", idx, ext)
			relativePath = filepath.ToSlash(path.Join("images", fileName))
			assetType = core.AssetTypeImage
		}

		assets[i].Type = assetType
		assets[i].FileName = fileName
		assets[i].RelativePath = relativePath
		assets[i].ContentType = contentType
		assets[i].Data = data
	}
	parsed.Assets = assets
	parsed.Images = imageAssetsFromAssets(assets)

	// Rewrite markdown links in a deterministic order.
	markdown := parsed.Markdown
	for i := range parsed.Assets {
		oldRef := parsed.Assets[i].SourceURL
		newRef := parsed.Assets[i].RelativePath
		markdown = strings.ReplaceAll(markdown, oldRef, newRef)
		if parsed.Assets[i].Type == core.AssetTypeVideo && videoMode == core.VideoModeEmbed {
			markdown = strings.ReplaceAll(markdown, "[video]("+newRef+")", videoMarkdownEmbed(newRef))
		}
	}
	parsed.Markdown = markdown

	if err := c.store.Save(ctx, parsed); err != nil {
		return nil, fmt.Errorf("save article: %w", err)
	}
	return parsed, nil
}

func videoMarkdownEmbed(src string) string {
	escaped := html.EscapeString(src)
	return "<video controls src=\"" + escaped + "\"></video>"
}

func (c *Converter) downloadAsset(ctx context.Context, assetURL string) ([]byte, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, "", "", err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", err
	}
	contentType := resp.Header.Get("Content-Type")
	ext := assetExt(contentType, assetURL)
	return data, contentType, ext, nil
}

func sanitizeFileName(s string) string {
	s = strings.TrimSpace(s)
	s = unsafeFileChars.ReplaceAllString(s, "_")
	s = spaceChars.ReplaceAllString(s, "_")
	s = strings.Trim(s, "._")
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

func assetExt(contentType, sourceURL string) string {
	if contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err == nil {
			if exts, ok := mime.ExtensionsByType(mediaType); ok == nil && len(exts) > 0 {
				return exts[0]
			}
		}
	}
	u, err := url.Parse(sourceURL)
	if err == nil {
		ext := strings.ToLower(path.Ext(u.Path))
		if ext != "" && len(ext) <= 5 {
			return ext
		}
	}
	if inferAssetType(contentType) == core.AssetTypeVideo {
		return ".mp4"
	}
	return ".jpg"
}

func inferAssetType(contentType string) core.AssetType {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	if strings.HasPrefix(mediaType, "video/") {
		return core.AssetTypeVideo
	}
	if strings.HasPrefix(mediaType, "image/") {
		return core.AssetTypeImage
	}
	return ""
}

func imageAssetsFromAssets(assets []core.Asset) []core.ImageAsset {
	images := make([]core.ImageAsset, 0, len(assets))
	for _, asset := range assets {
		if asset.Type != core.AssetTypeImage {
			continue
		}
		images = append(images, core.ImageAsset{
			SourceURL:    asset.SourceURL,
			RelativePath: asset.RelativePath,
			FileName:     asset.FileName,
			ContentType:  asset.ContentType,
			Data:         asset.Data,
		})
	}
	return images
}

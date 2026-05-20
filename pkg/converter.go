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
	"github.com/kbsink-org/kbsink/pkg/driver"
	"github.com/kbsink-org/kbsink/pkg/logger"
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
	driver core.Driver
	parser core.Parser
	store  core.Storage
	client core.HTTPClient
	log    logger.Logger
}

// NewConverter creates a converter with sensible defaults.
//
// By default it uses: HTMLDriver + LocalStorage + http.DefaultClient.
// Inject a Parser with WithParser (platform plugins live in kbsink-plugins).
// Use WithLogger and optional WithMinLevel for diagnostics.
func NewConverter(opts ...Option) *Converter {
	cfg := &converterConfig{
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	log := cfg.log
	if log != nil && cfg.minLevelSet {
		log = logger.WithMinLevel(log, cfg.minLevel)
	}

	if cfg.driver == nil {
		cfg.driver = driver.NewHTMLDriver(cfg.client, "", log)
	}
	if cfg.store == nil {
		cfg.store = stg.NewLocalStorage(defaultOutputRoot, log)
	}

	return &Converter{
		driver: cfg.driver,
		parser: cfg.parser,
		store:  cfg.store,
		client: cfg.client,
		log:    log,
	}
}

// Convert fetches article HTML, parses markdown, downloads assets, and saves output.
func (c *Converter) Convert(ctx context.Context, articleURL string, opts core.ConvertOptions) (*core.ArticleResult, error) {
	if articleURL == "" {
		return nil, fmt.Errorf("article url is required")
	}
	if c.parser == nil {
		return nil, fmt.Errorf("parser is required: use kbsink.WithParser (see kbsink-plugins)")
	}

	outputRoot := opts.OutputRoot
	if outputRoot == "" {
		outputRoot = defaultOutputRoot
	}
	videoMode := opts.VideoMode
	if videoMode == "" {
		videoMode = core.VideoModeLink
	}

	c.logInfo("convert start", "url", articleURL, "outputRoot", outputRoot, "videoMode", videoMode)

	raw, err := c.driver.Fetch(ctx, articleURL)
	if err != nil {
		c.logError("convert fetch failed", "url", articleURL, "err", err)
		return nil, fmt.Errorf("fetch article: %w", err)
	}

	outDir := outputRoot
	parsed, err := c.parser.Parse(ctx, raw, outDir)
	if err != nil {
		c.logError("convert parse failed", "url", articleURL, "err", err)
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

	c.logInfo("convert download assets", "count", len(assets))
	imageIdx := 0
	videoIdx := 0
	for i := range assets {
		c.logDebug("convert download asset", "index", i+1, "url", assets[i].SourceURL, "type", assets[i].Type)
		data, contentType, ext, dlErr := c.downloadAsset(ctx, assets[i].SourceURL)
		if dlErr != nil {
			c.logError("convert download asset failed", "url", assets[i].SourceURL, "err", dlErr)
			return nil, fmt.Errorf("download asset %q: %w", assets[i].SourceURL, dlErr)
		}
		assetType := assets[i].Type
		if assetType == "" {
			assetType = inferAssetType(contentType)
		}
		ext = assetExt(contentType, assets[i].SourceURL, assetType, ext)
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
		c.logDebug("convert asset stored", "file", fileName, "bytes", len(data), "contentType", contentType)
	}
	parsed.Assets = assets
	parsed.Images = imageAssetsFromAssets(assets)

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
		c.logError("convert save failed", "outputDir", parsed.OutputDir, "err", err)
		return nil, fmt.Errorf("save article: %w", err)
	}

	c.logInfo("convert done", "title", parsed.Title, "outputDir", parsed.OutputDir, "assets", len(parsed.Assets))
	return parsed, nil
}

func (c *Converter) logDebug(msg string, kv ...any) {
	if c.log != nil {
		c.log.Debug(msg, kv...)
	}
}

func (c *Converter) logInfo(msg string, kv ...any) {
	if c.log != nil {
		c.log.Info(msg, kv...)
	}
}

func (c *Converter) logWarn(msg string, kv ...any) {
	if c.log != nil {
		c.log.Warn(msg, kv...)
	}
}

func (c *Converter) logError(msg string, kv ...any) {
	if c.log != nil {
		c.log.Error(msg, kv...)
	}
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
		c.logWarn("asset download bad status", "url", assetURL, "status", resp.Status)
		return nil, "", "", fmt.Errorf("unexpected status: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", "", err
	}
	contentType := resp.Header.Get("Content-Type")
	ext := assetExt(contentType, assetURL, "", "")
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

func preferredImageExt(mediaType string) string {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}

func pickKnownGoodExt(exts []string) string {
	priority := []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg", ".bmp", ".ico"}
	lower := make([]string, len(exts))
	for i, e := range exts {
		lower[i] = strings.ToLower(e)
	}
	for _, want := range priority {
		for _, e := range lower {
			if e == want {
				return e
			}
		}
	}
	return exts[0]
}

func assetExt(contentType, sourceURL string, declared core.AssetType, fromMIME string) string {
	ext := fromMIME
	if ext == "" && contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err == nil {
			if e := preferredImageExt(mediaType); e != "" {
				ext = e
			} else if exts, extErr := mime.ExtensionsByType(mediaType); extErr == nil && len(exts) > 0 {
				ext = pickKnownGoodExt(exts)
			}
		}
	}
	if ext == "" {
		if u, err := url.Parse(sourceURL); err == nil {
			if e := strings.ToLower(path.Ext(u.Path)); e != "" && len(e) <= 5 {
				ext = e
			}
		}
	}
	if declared == core.AssetTypeVideo {
		if ext == "" || isImageFileExt(ext) {
			if inferAssetType(contentType) == core.AssetTypeVideo {
				return videoExtFromMIME(contentType)
			}
			return ".mp4"
		}
		return ext
	}
	if ext != "" {
		return ext
	}
	if inferAssetType(contentType) == core.AssetTypeVideo {
		return videoExtFromMIME(contentType)
	}
	return ".jpg"
}

func isImageFileExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg", ".bmp", ".ico":
		return true
	default:
		return false
	}
}

func videoExtFromMIME(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ".mp4"
	}
	if exts, extErr := mime.ExtensionsByType(mediaType); extErr == nil && len(exts) > 0 {
		for _, e := range exts {
			le := strings.ToLower(e)
			if le != ".jpg" && le != ".jpeg" && le != ".png" && !isImageFileExt(le) {
				return le
			}
		}
	}
	return ".mp4"
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

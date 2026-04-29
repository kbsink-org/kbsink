package parser

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd/core"
)

// XiaohongshuParser extracts metadata/content/media from Xiaohongshu share HTML.
type XiaohongshuParser struct{}

func NewXiaohongshuParser() *XiaohongshuParser {
	return &XiaohongshuParser{}
}

func (p *XiaohongshuParser) Parse(_ context.Context, fetched *core.FetchResult, outputDir string) (*core.ArticleResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fetched.HTML))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	title := firstNonEmpty(
		strings.TrimSpace(doc.Find(`meta[property="og:title"]`).AttrOr("content", "")),
		strings.TrimSpace(doc.Find("title").First().Text()),
	)
	account := strings.TrimSpace(doc.Find(`meta[name="author"]`).AttrOr("content", ""))
	if account == "" {
		account = extractJSONField(fetched.HTML, "nickname")
	}

	contentSel := firstSelection(doc,
		"#detail-desc",
		".note-content",
		".content",
		"article",
	)

	assets := collectMediaAssets(contentSel, fetched.HTML, outputDir)
	md := ""
	rawHTML := fetched.HTML
	if contentSel != nil && contentSel.Length() > 0 {
		md = SelectionToMarkdown(contentSel)
		if inner, htmlErr := contentSel.Html(); htmlErr == nil {
			rawHTML = inner
		}
	}

	md = strings.TrimSpace(md)
	if md == "" {
		md = fmt.Sprintf("# %s\n", title)
	}
	md += buildVideoLinksMarkdown(assets)

	return &core.ArticleResult{
		Title:          title,
		AccountName:    account,
		SourceURL:      fetched.URL,
		Markdown:       md,
		Assets:         assets,
		Images:         imageAssetsFromGenericAssets(assets, outputDir),
		RawHTMLContent: rawHTML,
	}, nil
}

func firstSelection(doc *goquery.Document, selectors ...string) *goquery.Selection {
	for _, selector := range selectors {
		sel := doc.Find(selector).First()
		if sel.Length() > 0 {
			return sel
		}
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func collectMediaAssets(contentSel *goquery.Selection, rawHTML string, outputDir string) []core.Asset {
	assets := make([]core.Asset, 0)
	seen := map[string]struct{}{}
	addAsset := func(assetType core.AssetType, src string) {
		src = strings.TrimSpace(src)
		if src == "" {
			return
		}
		if _, ok := seen[src]; ok {
			return
		}
		seen[src] = struct{}{}
		assets = append(assets, core.Asset{
			Type:         assetType,
			SourceURL:    src,
			RelativePath: path.Join(outputDir, "images"),
		})
	}

	if contentSel != nil {
		contentSel.Find("img").Each(func(_ int, sel *goquery.Selection) {
			src := strings.TrimSpace(sel.AttrOr("data-src", ""))
			if src == "" {
				src = strings.TrimSpace(sel.AttrOr("src", ""))
			}
			addAsset(core.AssetTypeImage, src)
		})
		contentSel.Find("video source, video").Each(func(_ int, sel *goquery.Selection) {
			src := strings.TrimSpace(sel.AttrOr("src", ""))
			addAsset(core.AssetTypeVideo, src)
		})
	}

	for _, src := range extractURLsByExtensions(rawHTML, []string{".mp4", ".mov", ".m4v"}) {
		addAsset(core.AssetTypeVideo, src)
	}
	for _, src := range extractURLsByExtensions(rawHTML, []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}) {
		addAsset(core.AssetTypeImage, src)
	}

	return assets
}

func extractURLsByExtensions(raw string, exts []string) []string {
	quotedURLPattern := regexp.MustCompile(`https?://[^"']+`)
	matches := quotedURLPattern.FindAllString(raw, -1)
	results := make([]string, 0)
	for _, match := range matches {
		lower := strings.ToLower(match)
		for _, ext := range exts {
			if strings.Contains(lower, ext) {
				results = append(results, strings.Split(match, `\u0026`)[0])
				break
			}
		}
	}
	return results
}

func buildVideoLinksMarkdown(assets []core.Asset) string {
	var b strings.Builder
	for _, asset := range assets {
		if asset.Type != core.AssetTypeVideo {
			continue
		}
		_, _ = b.WriteString("\n\n[video](" + asset.SourceURL + ")")
	}
	return b.String()
}

func extractJSONField(rawHTML, field string) string {
	pattern := regexp.MustCompile(`"` + regexp.QuoteMeta(field) + `"\s*:\s*"([^"]+)"`)
	match := pattern.FindStringSubmatch(rawHTML)
	if len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func imageAssetsFromGenericAssets(assets []core.Asset, outputDir string) []core.ImageAsset {
	images := make([]core.ImageAsset, 0, len(assets))
	for _, asset := range assets {
		if asset.Type != core.AssetTypeImage {
			continue
		}
		images = append(images, core.ImageAsset{
			SourceURL:    asset.SourceURL,
			RelativePath: path.Join(outputDir, "images"),
		})
	}
	return images
}

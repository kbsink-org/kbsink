package parser

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kbsink-org/kbsink/pkg/core"
)

// WechatParser extracts metadata/content from WeChat article HTML.
type WechatParser struct{}

func NewWechatParser() *WechatParser {
	return &WechatParser{}
}

func (p *WechatParser) Parse(_ context.Context, fetched *core.FetchResult, outputDir string) (*core.ArticleResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fetched.HTML))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	title := strings.TrimSpace(doc.Find("#activity-name").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").First().Text())
	}
	account := strings.TrimSpace(doc.Find("#js_name").First().Text())
	pubText := strings.TrimSpace(doc.Find("#publish_time").First().Text())
	sourceURL := strings.TrimSpace(doc.Find("#js_view_source").First().AttrOr("href", ""))
	if sourceURL == "" {
		sourceURL = fetched.URL
	}

	contentSel := doc.Find("#js_content").First()
	if contentSel.Length() == 0 {
		return nil, fmt.Errorf("cannot find #js_content")
	}

	images := make([]core.ImageAsset, 0)
	assets := make([]core.Asset, 0)
	contentSel.Find("img").Each(func(_ int, sel *goquery.Selection) {
		src := strings.TrimSpace(sel.AttrOr("data-src", ""))
		if src == "" {
			src = strings.TrimSpace(sel.AttrOr("src", ""))
		}
		if isFetchableAssetURL(src) {
			images = append(images, core.ImageAsset{
				SourceURL:    src,
				RelativePath: path.Join(outputDir, "images"),
			})
			assets = append(assets, core.Asset{
				Type:      core.AssetTypeImage,
				SourceURL: src,
			})
		}
	})

	md := SelectionToMarkdown(contentSel)
	rawHTML, _ := contentSel.Html()
	var publishedAt *time.Time
	if pubText != "" {
		if t, parseErr := parseWechatTime(pubText); parseErr == nil {
			publishedAt = &t
		}
	}

	return &core.ArticleResult{
		Title:          title,
		AccountName:    account,
		PublishedAt:    publishedAt,
		SourceURL:      sourceURL,
		Markdown:       md,
		Assets:         assets,
		Images:         images,
		RawHTMLContent: rawHTML,
	}, nil
}

func parseWechatTime(raw string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02",
	}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, raw, time.Local)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", raw)
}

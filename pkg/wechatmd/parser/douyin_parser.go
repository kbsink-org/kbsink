package parser

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd/core"
)

// DouyinParser extracts metadata/content/media from Douyin share HTML.
type DouyinParser struct{}

func NewDouyinParser() *DouyinParser {
	return &DouyinParser{}
}

func (p *DouyinParser) Parse(_ context.Context, fetched *core.FetchResult, outputDir string) (*core.ArticleResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fetched.HTML))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	title := firstNonEmpty(
		strings.TrimSpace(doc.Find(`meta[property="og:title"]`).AttrOr("content", "")),
		strings.TrimSpace(doc.Find("title").First().Text()),
		extractJSONField(fetched.HTML, "desc"),
	)
	account := firstNonEmpty(
		strings.TrimSpace(doc.Find(`meta[name="author"]`).AttrOr("content", "")),
		extractJSONField(fetched.HTML, "nickname"),
	)

	contentSel := firstSelection(doc,
		".video-desc",
		".detail-content",
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

	if strings.TrimSpace(md) == "" {
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

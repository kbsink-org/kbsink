package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd/core"
)

func TestXiaohongshuParser_Parse(t *testing.T) {
	html := `
<html>
<head>
  <meta property="og:title" content="XHS Note Title" />
  <meta name="author" content="xhs_author" />
</head>
<body>
  <article class="note-content">
    <p>Hello XHS</p>
    <img src="https://cdn.example.com/a.jpg" />
    <video src="https://cdn.example.com/v.mp4"></video>
  </article>
</body>
</html>`
	res, err := NewXiaohongshuParser().Parse(context.Background(), &core.FetchResult{
		URL:  "https://www.xiaohongshu.com/explore/abc",
		HTML: html,
	}, "output")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if res.Title != "XHS Note Title" {
		t.Fatalf("unexpected title: %q", res.Title)
	}
	if len(res.Assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(res.Assets))
	}
	if !strings.Contains(res.Markdown, "[video](https://cdn.example.com/v.mp4)") {
		t.Fatalf("markdown should include video link: %q", res.Markdown)
	}
}

func TestDouyinParser_Parse(t *testing.T) {
	html := `
<html>
<head>
  <meta property="og:title" content="Douyin Video Title" />
</head>
<body>
  <div class="detail-content">
    <p>Douyin Desc</p>
    <img src="https://cdn.example.com/dy.jpg" />
  </div>
  <script>window.__DATA__={"nickname":"dy_author","videoUrl":"https://cdn.example.com/dy.mp4"}</script>
</body>
</html>`
	res, err := NewDouyinParser().Parse(context.Background(), &core.FetchResult{
		URL:  "https://www.douyin.com/video/123",
		HTML: html,
	}, "output")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if res.Title != "Douyin Video Title" {
		t.Fatalf("unexpected title: %q", res.Title)
	}
	if res.AccountName != "dy_author" {
		t.Fatalf("unexpected account name: %q", res.AccountName)
	}
	if len(res.Assets) < 2 {
		t.Fatalf("expected at least 2 assets, got %d", len(res.Assets))
	}
	if !strings.Contains(res.Markdown, "[video](https://cdn.example.com/dy.mp4)") {
		t.Fatalf("markdown should include video link: %q", res.Markdown)
	}
}

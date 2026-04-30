package kbsink

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kbsink-org/kbsink/pkg/core"
	prs "github.com/kbsink-org/kbsink/pkg/parser"
)

type memoryStorage struct {
	saved *core.ArticleResult
}

func (m *memoryStorage) Save(_ context.Context, article *core.ArticleResult) error {
	m.saved = article
	return nil
}

func TestSanitizeFileName(t *testing.T) {
	got := sanitizeFileName(`  hello:/\*world?  `)
	if got != "hello____world" {
		t.Fatalf("unexpected sanitized file name: %q", got)
	}
}

func TestConvertWithDefaultHTMLDriverAndParser(t *testing.T) {
	imageData := []byte{1, 2, 3}
	mux := http.NewServeMux()
	mux.HandleFunc("/article", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`
<html>
  <head><title>fallback-title</title></head>
  <body>
    <h1 id="activity-name">Test Article</h1>
    <strong id="js_name">Test Account</strong>
    <em id="publish_time">2026-04-28 10:20:30</em>
    <div id="js_content">
      <p>Hello</p>
      <img data-src="` + "http://example.invalid/img.jpg" + `" />
    </div>
  </body>
</html>`))
	})
	mux.HandleFunc("/img.jpg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(imageData)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Ensure image src points to our test server.
	articleURL := ts.URL + "/article"
	html := `
<html><body>
<h1 id="activity-name">Test Article</h1>
<strong id="js_name">Test Account</strong>
<em id="publish_time">2026-04-28 10:20:30</em>
<div id="js_content"><p>Hello</p><img data-src="` + ts.URL + `/img.jpg" /></div>
</body></html>`
	driver := &stubDriver{res: &core.FetchResult{URL: articleURL, HTML: html}}
	memStore := &memoryStorage{}

	c := NewConverter(
		WithDriver(driver),
		WithParser(prs.NewWechatParser()),
		WithStorage(memStore),
		WithHTTPClient(ts.Client()),
	)
	res, err := c.Convert(context.Background(), articleURL, core.ConvertOptions{OutputRoot: "output"})
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}

	if res.Title != "Test Article" {
		t.Fatalf("unexpected title: %q", res.Title)
	}
	if len(res.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(res.Images))
	}
	if !strings.Contains(res.Markdown, "images/img_001.") {
		t.Fatalf("markdown image path not rewritten: %q", res.Markdown)
	}
	if memStore.saved == nil {
		t.Fatalf("storage save should be called")
	}
}

type stubDriver struct {
	res *core.FetchResult
	err error
}

func (s *stubDriver) Fetch(_ context.Context, _ string) (*core.FetchResult, error) {
	return s.res, s.err
}

type stubParser struct {
	res *core.ArticleResult
	err error
}

func (s *stubParser) Parse(_ context.Context, _ *core.FetchResult, _ string) (*core.ArticleResult, error) {
	return s.res, s.err
}

func TestConvertWithImageAndVideoAssets(t *testing.T) {
	imageData := []byte{1, 2, 3}
	videoData := []byte{4, 5, 6}
	mux := http.NewServeMux()
	mux.HandleFunc("/img.jpg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(imageData)
	})
	mux.HandleFunc("/video.mp4", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write(videoData)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	parser := &stubParser{
		res: &core.ArticleResult{
			Title:    "Mixed Media",
			Markdown: "![i](" + ts.URL + "/img.jpg)\n\n[video](" + ts.URL + "/video.mp4)",
			Assets: []core.Asset{
				{Type: core.AssetTypeImage, SourceURL: ts.URL + "/img.jpg"},
				{Type: core.AssetTypeVideo, SourceURL: ts.URL + "/video.mp4"},
			},
		},
	}
	memStore := &memoryStorage{}
	c := NewConverter(
		WithDriver(&stubDriver{res: &core.FetchResult{URL: ts.URL + "/post", HTML: "<html></html>"}}),
		WithParser(parser),
		WithStorage(memStore),
		WithHTTPClient(ts.Client()),
	)
	res, err := c.Convert(context.Background(), ts.URL+"/post", core.ConvertOptions{OutputRoot: "output"})
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}
	if len(res.Assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(res.Assets))
	}
	if !strings.Contains(res.Markdown, "images/img_001.") {
		t.Fatalf("image path not rewritten: %q", res.Markdown)
	}
	if !strings.Contains(res.Markdown, "videos/video_001.") {
		t.Fatalf("video path not rewritten: %q", res.Markdown)
	}
	if len(res.Images) != 1 {
		t.Fatalf("expected 1 image in compatibility field, got %d", len(res.Images))
	}
}

func TestConvertWithEmbedVideoMode(t *testing.T) {
	videoData := []byte{4, 5, 6}
	mux := http.NewServeMux()
	mux.HandleFunc("/video.mp4", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write(videoData)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	parser := &stubParser{
		res: &core.ArticleResult{
			Title:    "Video Only",
			Markdown: "[video](" + ts.URL + "/video.mp4)",
			Assets: []core.Asset{
				{Type: core.AssetTypeVideo, SourceURL: ts.URL + "/video.mp4"},
			},
		},
	}
	memStore := &memoryStorage{}
	c := NewConverter(
		WithDriver(&stubDriver{res: &core.FetchResult{URL: ts.URL + "/post", HTML: "<html></html>"}}),
		WithParser(parser),
		WithStorage(memStore),
		WithHTTPClient(ts.Client()),
	)
	res, err := c.Convert(context.Background(), ts.URL+"/post", core.ConvertOptions{
		OutputRoot: "output",
		VideoMode:  core.VideoModeEmbed,
	})
	if err != nil {
		t.Fatalf("convert error: %v", err)
	}
	if !strings.Contains(res.Markdown, "<video controls src=\"videos/video_001.") {
		t.Fatalf("video should be embedded: %q", res.Markdown)
	}
}

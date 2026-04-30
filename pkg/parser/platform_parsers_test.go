package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kbsink-org/kbsink/pkg/core"
)

func TestXiaohongshuParser_Parse(t *testing.T) {
	html := mustReadTestHTML(t, "xiaohongshu_real.html")
	res, err := NewXiaohongshuParser().Parse(context.Background(), &core.FetchResult{
		URL:  "https://www.xiaohongshu.com/explore/69eca7e800000000230072ba",
		HTML: html,
	}, "output")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if strings.TrimSpace(res.Title) == "" {
		t.Fatalf("expected non-empty title from real snapshot")
	}
	if strings.TrimSpace(res.Markdown) == "" {
		t.Fatalf("expected non-empty markdown from real snapshot")
	}
}

func TestWechatParser_RealSnapshot(t *testing.T) {
	html := mustReadTestHTML(t, "wechat_real.html")
	res, err := NewWechatParser().Parse(context.Background(), &core.FetchResult{
		URL:  "https://mp.weixin.qq.com/s/Y7dyRC7CJ09miHWU6LBzBA",
		HTML: html,
	}, "output")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if strings.TrimSpace(res.Title) == "" {
		t.Fatalf("expected non-empty title from real snapshot")
	}
	if strings.TrimSpace(res.Markdown) == "" {
		t.Fatalf("expected non-empty markdown from real snapshot")
	}
}

func mustReadTestHTML(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join("testdata", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read testdata %q: %v", p, err)
	}
	return string(b)
}

package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd"
	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd/core"
	prs "github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd/parser"
)

func main() {
	var (
		outputRoot = flag.String("o", "output", "output root directory")
		timeout    = flag.Duration("timeout", 60*time.Second, "timeout for the conversion")
		printOnly  = flag.Bool("print", false, "print markdown to stdout (still downloads images; does not save files)")
		platform   = flag.String("platform", "auto", "platform parser: auto|wechat|xiaohongshu|douyin")
	)
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  %s [flags] <wechat-article-url>\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	articleURL := flag.Arg(0)

	ctx := context.Background()
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	parser, err := resolveParser(*platform, articleURL)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "resolve platform parser failed: %v\n", err)
		os.Exit(1)
	}
	converter := wechatmd.NewConverter(wechatmd.WithParser(parser))
	res, err := converter.Convert(ctx, articleURL, core.ConvertOptions{OutputRoot: *outputRoot})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "convert failed: %v\n", err)
		os.Exit(1)
	}

	if *printOnly {
		_, _ = fmt.Fprint(os.Stdout, res.Markdown)
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "title: %s\n", res.Title)
	_, _ = fmt.Fprintf(os.Stdout, "markdown: %s\n", res.MarkdownPath)
	_, _ = fmt.Fprintf(os.Stdout, "images: %d\n", len(res.Images))
	videoCount := 0
	for _, asset := range res.Assets {
		if asset.Type == core.AssetTypeVideo {
			videoCount++
		}
	}
	_, _ = fmt.Fprintf(os.Stdout, "videos: %d\n", videoCount)
}

func resolveParser(platform, articleURL string) (prs.Parser, error) {
	selected := strings.TrimSpace(strings.ToLower(platform))
	if selected == "" {
		selected = "auto"
	}
	switch selected {
	case "wechat":
		return prs.NewWechatParser(), nil
	case "xiaohongshu", "xhs":
		return prs.NewXiaohongshuParser(), nil
	case "douyin":
		return prs.NewDouyinParser(), nil
	case "auto":
		host := strings.ToLower(strings.TrimSpace(articleURLHost(articleURL)))
		switch {
		case strings.Contains(host, "xiaohongshu.com"), strings.Contains(host, "xhslink.com"):
			return prs.NewXiaohongshuParser(), nil
		case strings.Contains(host, "douyin.com"), strings.Contains(host, "iesdouyin.com"):
			return prs.NewDouyinParser(), nil
		case strings.Contains(host, "weixin.qq.com"), strings.Contains(host, "mp.weixin.qq.com"):
			return prs.NewWechatParser(), nil
		default:
			return nil, fmt.Errorf("unsupported host for auto platform selection: %q", host)
		}
	default:
		return nil, fmt.Errorf("unsupported platform %q, expected auto|wechat|xiaohongshu|douyin", platform)
	}
}

func articleURLHost(articleURL string) string {
	u, err := url.Parse(articleURL)
	if err != nil {
		return ""
	}
	return u.Host
}

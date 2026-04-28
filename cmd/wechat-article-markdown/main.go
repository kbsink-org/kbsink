package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd"
	"github.com/SolaTyolo/wechat-article-markdown/pkg/wechatmd/core"
)

func main() {
	var (
		outputRoot = flag.String("o", "output", "output root directory")
		timeout    = flag.Duration("timeout", 60*time.Second, "timeout for the conversion")
		printOnly  = flag.Bool("print", false, "print markdown to stdout (still downloads images; does not save files)")
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

	converter := wechatmd.NewConverter()
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
}

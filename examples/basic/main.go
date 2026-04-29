package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kbsink-org/kbsink/pkg"
	"github.com/kbsink-org/kbsink/pkg/core"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./examples/basic <wechat-article-url>")
	}
	url := os.Args[1]

	converter := kbsink.NewConverter()
	res, err := converter.Convert(context.Background(), url, core.ConvertOptions{
		OutputRoot: "output",
	})
	if err != nil {
		log.Fatalf("convert failed: %v", err)
	}

	fmt.Printf("title: %s\n", res.Title)
	fmt.Printf("markdown: %s\n", res.MarkdownPath)
	fmt.Printf("images: %d\n", len(res.Images))
}

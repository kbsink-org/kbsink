# kbsink (Go Library)

[中文文档](./README-zh.md)

A reusable Go library that converts WeChat Official Account article URLs into structured markdown output, with pluggable storage and driver interfaces.

## Table of Contents

- [Features](#features)
- [Install](#install)
- [Quick Start](#quick-start)
- [Extensibility](#extensibility)
  - [Custom Driver](#custom-driver)
  - [Custom Storage (e.g. S3)](#custom-storage-eg-s3)
- [Example](#example)

## Features

- Unified entrypoint: `Convert(ctx, url, opts)` via `Converter`
- Pluggable `Driver` interface (official API, anti-bot browser, HTML fetcher, etc.)
- Default `HTMLDriver` for direct HTTP fetching
- Pluggable `Parser` with default `WechatParser` (metadata extraction + markdown conversion)
- Pluggable `Storage` interface with default `LocalStorage`

## Install

```bash
go get github.com/kbsink-org/kbsink
```

## Quick Start

```go
converter := kbsink.NewConverter()
res, err := converter.Convert(ctx, "https://mp.weixin.qq.com/s/xxxx", core.ConvertOptions{
    OutputRoot: "output",
})
```

Default output structure:

```text
output/
└── <article-title>/
    ├── <article-title>.md
    └── images/
        ├── img_001.png
        ├── img_002.png
        └── ...
```

## Extensibility

### Custom Driver

Implement and inject a `Driver`:

```go
type MyDriver struct{}
func (d *MyDriver) Fetch(ctx context.Context, url string) (*core.FetchResult, error) {
    // Official API / anti-bot browser / any custom fetch logic
}
```

### Custom Storage (e.g. S3)

Implement and inject a `Storage`:

```go
type S3Storage struct{}
func (s *S3Storage) Save(ctx context.Context, article *core.ArticleResult) error {
    // Upload article.Markdown and article.Images
}
```

## Example

```bash
go run ./examples/basic "https://mp.weixin.qq.com/s/xxxx"
```

## CLI

Build/install locally:

```bash
go install ./cmd/kb-sink-md
```

Run:

```bash
kb-sink-md -o output "https://mp.weixin.qq.com/s/xxxx"
```


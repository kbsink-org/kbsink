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
    VideoMode:  core.VideoModeEmbed, // optional: core.VideoModeLink (default)
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

## CLI

Build/install locally:

```bash
go install ./cmd/kb-sink-md
```

Run:

```bash
kb-sink-md --plugin wechat -o output --video-mode embed "https://mp.weixin.qq.com/s/xxxx"
```

For **Douyin** (`--plugin douyin`), keep using **`kb-sink-md`** from this repository and add **`kbsink-plugin-douyin`** to your `PATH` (releases or `go build` from [douyin-plugin](https://github.com/kbsink-org/douyin-plugin), `./cmd/kbsink-plugin-douyin`). The core `kb-sink-md` binary registers only **wechat** and **xhs** in-process; Douyin runs as an external subprocess via the stdin/stdout JSON protocol (`pkg/pluginexec`). You may also set **`KBSINK_PLUGIN_DOUYIN`** to the plugin binary path instead of relying on `PATH`.


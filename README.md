# kbsink (Go Library)

[中文文档](./README-zh.md)

A reusable Go library for converting article URLs into structured markdown: pluggable **Driver**, **Parser**, and **Storage** with a unified `Converter` pipeline.

Platform-specific parsers (WeChat, Xiaohongshu, Douyin, …) live in **[kbsink-plugins](https://github.com/kbsink-org/kbsink-plugins)**.

## Features

- Unified entrypoint: `Convert(ctx, url, opts)` via `Converter`
- Pluggable `Driver` (HTTP HTML fetch, custom APIs, etc.)
- Pluggable `Parser` (inject from kbsink-plugins or your own)
- Pluggable `Storage` with default `LocalStorage`
- Optional `PrepareAssetRequest` for CDN-specific download headers
- Injectable `logger.Logger` with levels (`debug` / `info` / `warn` / `error`) on converter, driver, parser, and storage

## Install

```bash
go get github.com/kbsink-org/kbsink
go get github.com/kbsink-org/kbsink-plugins
```

## Quick Start

```go
import (
	kbsink "github.com/kbsink-org/kbsink/pkg"
	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/logger"
	"github.com/kbsink-org/kbsink-plugins/pkg/wechat"
)

converter := kbsink.NewConverter(
	kbsink.WithParser(wechat.NewParser()),
	kbsink.WithDriver(wechat.NewDriver(nil)),
	kbsink.WithLogger(logger.Std()),
	kbsink.WithMinLevel(logger.LevelInfo),
)
res, err := converter.Convert(ctx, "https://mp.weixin.qq.com/s/xxxx", core.ConvertOptions{
	OutputRoot: "output",
})
```

## Command-line tool

Use **[kbsink-cli](https://github.com/kbsink-org/kbsink-cli)** for a `kbsink` binary with plugins pre-wired (`--plugin` or URL auto-detect).

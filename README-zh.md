# kbsink（Go 库）

[English README](./README.md)

可复用的 Go 库：通过可插拔的 **Driver**、**Parser**、**Storage** 与统一的 `Converter` 流水线，将文章 URL 转为结构化 Markdown。

各平台解析器（微信、小红书、抖音等）在 **[kbsink-plugins](https://github.com/kbsink-org/kbsink-plugins)** 仓库中维护。

## 功能特性

- 统一入口：`Converter.Convert(ctx, url, opts)`
- 可插拔 `Driver`（HTTP HTML 抓取、自定义 API 等）
- 可插拔 `Parser`（从 kbsink-plugins 注入或自行实现）
- 可插拔 `Storage`，默认 `LocalStorage`
- 可选 `PrepareAssetRequest`，用于 CDN 资源下载请求头

## 安装

```bash
go get github.com/kbsink-org/kbsink
go get github.com/kbsink-org/kbsink-plugins
```

## 快速开始

```go
import (
	kbsink "github.com/kbsink-org/kbsink/pkg"
	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink-plugins/pkg/wechat"
)

converter := kbsink.NewConverter(
	kbsink.WithParser(wechat.NewParser()),
	kbsink.WithDriver(wechat.NewDriver(nil)),
)
res, err := converter.Convert(ctx, "https://mp.weixin.qq.com/s/xxxx", core.ConvertOptions{
	OutputRoot: "output",
})
```

## 命令行工具

使用 **[kbsink-cli](https://github.com/kbsink-org/kbsink-cli)** 获取已集成插件的 `kbsink` 二进制（`--plugin` 或根据 URL 自动识别）。

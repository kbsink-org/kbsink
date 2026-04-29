# kbsink（Go 库）

[English README](./README.md)

一个可复用的 Go 库：输入微信公众号文章 URL，输出结构化 Markdown 结果，并支持可插拔的存储与抓取驱动接口。

## 目录

- [功能特性](#功能特性)
- [安装](#安装)
- [快速开始](#快速开始)
- [可扩展性](#可扩展性)
  - [自定义 Driver](#自定义-driver)
  - [自定义 Storage（例如 S3）](#自定义-storage例如-s3)
- [示例](#示例)

## 功能特性

- 统一入口：通过 `Converter` 调用 `Convert(ctx, url, opts)`
- 可插拔 `Driver` 接口（官方接口、反扒浏览器、HTML 抓取等）
- 默认 `HTMLDriver`，直接 HTTP 拉取页面
- 可插拔 `Parser`，默认 `WechatParser`（提取元数据并转 Markdown）
- 可插拔 `Storage` 接口，默认 `LocalStorage`

## 安装

```bash
go get github.com/kbsink-org/kbsink
```

## 快速开始

```go
converter := kbsink.NewConverter()
res, err := converter.Convert(ctx, "https://mp.weixin.qq.com/s/xxxx", core.ConvertOptions{
    OutputRoot: "output",
})
```

默认输出结构：

```text
output/
└── <article-title>/
    ├── <article-title>.md
    └── images/
        ├── img_001.png
        ├── img_002.png
        └── ...
```

## 可扩展性

### 自定义 Driver

实现并注入 `Driver`：

```go
type MyDriver struct{}
func (d *MyDriver) Fetch(ctx context.Context, url string) (*core.FetchResult, error) {
    // 官方接口 / 反扒浏览器 / 其他抓取方式
}
```

### 自定义 Storage（例如 S3）

实现并注入 `Storage`：

```go
type S3Storage struct{}
func (s *S3Storage) Save(ctx context.Context, article *core.ArticleResult) error {
    // 上传 article.Markdown 和 article.Images
}
```

## 示例

```bash
go run ./examples/basic "https://mp.weixin.qq.com/s/xxxx"
```

## CLI

本地安装/构建：

```bash
go install ./cmd/kb-sink-md
```

运行：

```bash
kb-sink-md -o output "https://mp.weixin.qq.com/s/xxxx"
```


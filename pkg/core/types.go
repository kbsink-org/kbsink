package core

import "time"

// FetchResult is the raw payload returned by a core.Driver.
type FetchResult struct {
	URL  string
	HTML string
}

// ImageAsset stores an image mapping and optional binary payload.
type ImageAsset struct {
	SourceURL    string
	RelativePath string
	FileName     string
	ContentType  string
	Data         []byte
}

// AssetType indicates the media category of one asset.
type AssetType string

const (
	AssetTypeImage AssetType = "image"
	AssetTypeVideo AssetType = "video"
)

// Asset stores a generic media mapping and optional binary payload.
type Asset struct {
	Type         AssetType
	SourceURL    string
	RelativePath string
	FileName     string
	ContentType  string
	Data         []byte
}

// ArticleResult is the final structured output for one URL conversion.
type ArticleResult struct {
	Title          string
	SafeTitle      string
	AccountName    string
	PublishedAt    *time.Time
	SourceURL      string
	OutputDir      string
	MarkdownPath   string
	Markdown       string
	Assets         []Asset
	Images         []ImageAsset
	RawHTMLContent string
}

// ConvertOptions controls per-call conversion behavior.
type ConvertOptions struct {
	OutputRoot string
	VideoMode  VideoMode
}

// VideoMode controls how video assets are rendered in markdown.
type VideoMode string

const (
	VideoModeLink  VideoMode = "link"
	VideoModeEmbed VideoMode = "embed"
)

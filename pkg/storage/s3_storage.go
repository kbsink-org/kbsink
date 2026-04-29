package storage

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/kbsink-org/kbsink/pkg/core"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3PutObjectAPI interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// S3Storage saves markdown and images into S3-compatible object storage.
type S3Storage struct {
	client s3PutObjectAPI
	bucket string
	prefix string
}

func NewS3Storage(client *s3.Client, bucket, prefix string) (*S3Storage, error) {
	if client == nil {
		return nil, fmt.Errorf("s3 client is nil")
	}
	if strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("bucket is required")
	}
	prefix = normalizeS3Key(prefix)
	return &S3Storage{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}, nil
}

func (s *S3Storage) Save(ctx context.Context, article *core.ArticleResult) error {
	if article == nil {
		return fmt.Errorf("article is nil")
	}

	mdKey := s.fullKey(article.MarkdownPath)
	if err := s.put(ctx, mdKey, []byte(article.Markdown), "text/markdown; charset=utf-8"); err != nil {
		return fmt.Errorf("upload markdown: %w", err)
	}

	assets := article.Assets
	if len(assets) == 0 && len(article.Images) > 0 {
		assets = make([]core.Asset, 0, len(article.Images))
		for _, img := range article.Images {
			assets = append(assets, core.Asset{
				Type:         core.AssetTypeImage,
				SourceURL:    img.SourceURL,
				RelativePath: img.RelativePath,
				FileName:     img.FileName,
				ContentType:  img.ContentType,
				Data:         img.Data,
			})
		}
	}

	for _, asset := range assets {
		key := s.fullKey(path.Join(article.OutputDir, asset.RelativePath))
		contentType := strings.TrimSpace(asset.ContentType)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if err := s.put(ctx, key, asset.Data, contentType); err != nil {
			return fmt.Errorf("upload asset %q: %w", asset.FileName, err)
		}
	}
	return nil
}

func (s *S3Storage) put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *S3Storage) fullKey(rel string) string {
	key := normalizeS3Key(rel)
	if s.prefix == "" {
		return key
	}
	return path.Join(s.prefix, key)
}

func normalizeS3Key(raw string) string {
	raw = strings.ReplaceAll(raw, "\\", "/")
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "/")
	raw = path.Clean(raw)
	if raw == "." {
		return ""
	}
	return raw
}

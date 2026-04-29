package storage

import (
	"context"
	"io"
	"testing"

	"github.com/kbsink-org/kbsink/pkg/core"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type fakeS3Client struct {
	keys []string
}

func (f *fakeS3Client) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.keys = append(f.keys, *params.Key)
	if params.Body != nil {
		_, _ = io.ReadAll(params.Body)
	}
	return &s3.PutObjectOutput{}, nil
}

func TestS3StorageSave(t *testing.T) {
	client := &fakeS3Client{}
	s := &S3Storage{
		client: client,
		bucket: "demo-bucket",
		prefix: "articles",
	}
	article := &core.ArticleResult{
		OutputDir:    "output/demo",
		MarkdownPath: "output/demo/demo.md",
		Markdown:     "# demo",
		Images: []core.ImageAsset{
			{
				FileName:     "img_001.png",
				RelativePath: "images/img_001.png",
				ContentType:  "image/png",
				Data:         []byte{1, 2, 3},
			},
		},
	}

	if err := s.Save(context.Background(), article); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if len(client.keys) != 2 {
		t.Fatalf("expected 2 uploads, got %d", len(client.keys))
	}
	if client.keys[0] != "articles/output/demo/demo.md" {
		t.Fatalf("unexpected markdown key: %s", client.keys[0])
	}
	if client.keys[1] != "articles/output/demo/images/img_001.png" {
		t.Fatalf("unexpected image key: %s", client.keys[1])
	}
}

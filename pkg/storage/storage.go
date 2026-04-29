package storage

import (
	"context"

	"github.com/kbsink-org/kbsink/pkg/core"
)

// Storage persists article markdown and image assets.
type Storage interface {
	Save(ctx context.Context, article *core.ArticleResult) error
}

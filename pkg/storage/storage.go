// Package storage defines constructors for kbsink persistence. Types live in core (Storage interface)
// and internal/store; the default Converter uses local disk via NewLocalStorage.
package storage

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kbsink-org/kbsink/internal/store"
	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/logger"
)

// NewLocalStorage writes output under a filesystem root directory.
// log may be nil (uses [logger.Resolve]).
func NewLocalStorage(root string, log logger.Logger) core.Storage {
	return store.NewLocalStorage(root, log)
}

// NewS3Storage uploads markdown and assets to an S3-compatible bucket.
// log may be nil (uses [logger.Resolve]).
func NewS3Storage(client *s3.Client, bucket, prefix string, log logger.Logger) (core.Storage, error) {
	return store.NewS3Storage(client, bucket, prefix, log)
}

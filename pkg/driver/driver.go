package driver

import (
	"context"

	"github.com/kbsink-org/kbsink/pkg/core"
)

// Driver fetches article page source by URL.
type Driver interface {
	Fetch(ctx context.Context, url string) (*core.FetchResult, error)
}

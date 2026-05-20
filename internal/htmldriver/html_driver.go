package htmldriver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/logger"
)

// HTMLDriver fetches HTML from a page URL through plain HTTP.
type HTMLDriver struct {
	client    *http.Client
	userAgent string
	log       logger.Logger
}

// New returns an HTML fetch driver. userAgent must be non-empty (callers typically normalize empty to a default).
func New(client *http.Client, userAgent string, log logger.Logger) *HTMLDriver {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTMLDriver{
		client:    client,
		userAgent: userAgent,
		log:       log,
	}
}

func (d *HTMLDriver) Fetch(ctx context.Context, rawURL string) (*core.FetchResult, error) {
	if strings.TrimSpace(rawURL) == "" {
		return nil, core.NewCodedError(core.ErrCodeInvalidArgument, "url is required", nil)
	}

	if d.log != nil {
		d.log.Debug("html driver: fetch start", "url", rawURL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		if d.log != nil {
			d.log.Error("html driver: build request failed", "url", rawURL, "err", err)
		}
		return nil, core.NewCodedError(core.ErrCodeDriverBuildRequest, "build request", err)
	}
	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		if d.log != nil {
			d.log.Error("html driver: request failed", "url", rawURL, "err", err)
		}
		return nil, core.NewCodedError(core.ErrCodeDriverRequestFailed, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if d.log != nil {
			d.log.Warn("html driver: unexpected status", "url", rawURL, "status", resp.Status)
		}
		return nil, core.NewCodedError(
			core.ErrCodeDriverUnexpectedHTTP,
			fmt.Sprintf("unexpected status: %s", resp.Status),
			nil,
		)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if d.log != nil {
			d.log.Error("html driver: read body failed", "url", rawURL, "err", err)
		}
		return nil, core.NewCodedError(core.ErrCodeDriverReadBodyFailed, "read response body", err)
	}
	finalURL := rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	if d.log != nil {
		d.log.Info("html driver: fetch done", "url", finalURL, "htmlLen", len(body))
	}
	return &core.FetchResult{
		URL:  finalURL,
		HTML: string(body),
	}, nil
}

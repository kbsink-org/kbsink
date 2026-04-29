package driver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kbsink-org/kbsink/pkg/core"
)

// HTMLDriver fetches HTML from a page URL through plain HTTP.
type HTMLDriver struct {
	client *http.Client
}

func NewHTMLDriver(client *http.Client) *HTMLDriver {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTMLDriver{client: client}
}

func (d *HTMLDriver) Fetch(ctx context.Context, rawURL string) (*core.FetchResult, error) {
	if strings.TrimSpace(rawURL) == "" {
		return nil, core.NewCodedError(core.ErrCodeInvalidArgument, "url is required", nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, core.NewCodedError(core.ErrCodeDriverBuildRequest, "build request", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; wechatmd/1.0)")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, core.NewCodedError(core.ErrCodeDriverRequestFailed, "execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, core.NewCodedError(
			core.ErrCodeDriverUnexpectedHTTP,
			fmt.Sprintf("unexpected status: %s", resp.Status),
			nil,
		)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, core.NewCodedError(core.ErrCodeDriverReadBodyFailed, "read response body", err)
	}
	return &core.FetchResult{
		URL:  resp.Request.URL.String(),
		HTML: string(body),
	}, nil
}

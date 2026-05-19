package driver

import (
	"net/http"
	"strings"

	"github.com/kbsink-org/kbsink/internal/htmldriver"
	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/logger"
)

const defaultUserAgent = "Mozilla/5.0 (compatible; kbsink/1.0)"

// NewHTMLDriver returns an HTTP HTML fetch driver. If userAgent is empty or whitespace, defaultUserAgent is used.
// log may be nil (uses [logger.Resolve]).
func NewHTMLDriver(client *http.Client, userAgent string, log logger.Logger) core.Driver {
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		ua = defaultUserAgent
	}
	return htmldriver.New(client, ua, log)
}

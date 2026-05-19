package htmldriver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kbsink-org/kbsink/internal/htmldriver"
	"github.com/kbsink-org/kbsink/pkg/core"
)

func TestHTMLDriverFetch_ErrorCodes(t *testing.T) {
	d := htmldriver.New(nil, "Mozilla/5.0 (test)", nil)

	_, err := d.Fetch(context.Background(), "")
	if got := core.ErrorCodeOf(err); got != core.ErrCodeInvalidArgument {
		t.Fatalf("expected %s, got %s (err=%v)", core.ErrCodeInvalidArgument, got, err)
	}

	_, err = d.Fetch(context.Background(), "://bad-url")
	if got := core.ErrorCodeOf(err); got != core.ErrCodeDriverBuildRequest {
		t.Fatalf("expected %s, got %s (err=%v)", core.ErrCodeDriverBuildRequest, got, err)
	}
}

func TestHTMLDriverFetch_UnexpectedStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	d := htmldriver.New(ts.Client(), "Mozilla/5.0 (test)", nil)
	_, err := d.Fetch(context.Background(), ts.URL)
	if got := core.ErrorCodeOf(err); got != core.ErrCodeDriverUnexpectedHTTP {
		t.Fatalf("expected %s, got %s (err=%v)", core.ErrCodeDriverUnexpectedHTTP, got, err)
	}
}

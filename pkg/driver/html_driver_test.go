package driver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kbsink-org/kbsink/pkg/core"
	"github.com/kbsink-org/kbsink/pkg/driver"
)

func TestNewHTMLDriverFetch_ErrorCodes(t *testing.T) {
	d := driver.NewHTMLDriver(nil, "", nil)

	_, err := d.Fetch(context.Background(), "")
	if got := core.ErrorCodeOf(err); got != core.ErrCodeInvalidArgument {
		t.Fatalf("expected %s, got %s (err=%v)", core.ErrCodeInvalidArgument, got, err)
	}

	_, err = d.Fetch(context.Background(), "://bad-url")
	if got := core.ErrorCodeOf(err); got != core.ErrCodeDriverBuildRequest {
		t.Fatalf("expected %s, got %s (err=%v)", core.ErrCodeDriverBuildRequest, got, err)
	}
}

func TestNewHTMLDriverFetch_UnexpectedStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	d := driver.NewHTMLDriver(ts.Client(), "", nil)
	_, err := d.Fetch(context.Background(), ts.URL)
	if got := core.ErrorCodeOf(err); got != core.ErrCodeDriverUnexpectedHTTP {
		t.Fatalf("expected %s, got %s (err=%v)", core.ErrCodeDriverUnexpectedHTTP, got, err)
	}
}

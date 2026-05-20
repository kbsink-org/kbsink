package core

import "net/http"

// HTTPClient executes HTTP requests. Typically *http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

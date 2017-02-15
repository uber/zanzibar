package barClient

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/uber/zanzibar/runtime"
)

// BarClient is the http client for service Bar.
type BarClient zanzibar.HTTPClient

// NewClient returns a new http client for service Bar.
func NewClient(opts *zanzibar.HTTPClientOptions) *BarClient {
	baseURL := "http://" + opts.IP + ":" + strconv.Itoa(int(opts.Port))
	return &BarClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
		},
		BaseURL: baseURL,
	}
}

// ArgNotStruct calls "/arg-not-struct-path" endpoint.
func (c *BarClient) ArgNotStruct(ctx context.Context, r *ArgNotStructHTTPRequest, h http.Header) (*http.Response, error) {
	// Generate full URL.
	fullURL := c.BaseURL + "/arg-not-struct-path"

	rawBody, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	if h != nil {
		req.Header = h
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req.WithContext(ctx))
}

// MissingArg calls "/missing-arg-path" endpoint.
func (c *BarClient) MissingArg(ctx context.Context, h http.Header) (*http.Response, error) {
	// Generate full URL.
	fullURL := c.BaseURL + "/missing-arg-path"

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	if h != nil {
		req.Header = h
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req.WithContext(ctx))
}

// NoRequest calls "/no-request-path" endpoint.
func (c *BarClient) NoRequest(ctx context.Context, h http.Header) (*http.Response, error) {
	// Generate full URL.
	fullURL := c.BaseURL + "/no-request-path"

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	if h != nil {
		req.Header = h
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req.WithContext(ctx))
}

// Normal calls "/bar-path" endpoint.
func (c *BarClient) Normal(ctx context.Context, r *NormalHTTPRequest, h http.Header) (*http.Response, error) {
	// Generate full URL.
	fullURL := c.BaseURL + "/bar-path"

	rawBody, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	if h != nil {
		req.Header = h
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req.WithContext(ctx))
}

// TooManyArgs calls "/too-many-args-path" endpoint.
func (c *BarClient) TooManyArgs(ctx context.Context, r *TooManyArgsHTTPRequest, h http.Header) (*http.Response, error) {
	// Generate full URL.
	fullURL := c.BaseURL + "/too-many-args-path"

	rawBody, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fullURL, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	if h != nil {
		req.Header = h
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Client.Do(req.WithContext(ctx))
}

package barClient

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/uber/zanzibar/lib/http_client"
)

// BarClient is the http client for service Bar.
type BarClient httpClient.HTTPClient

// NewClient returns a new http client for service Bar.
func NewClient(opts *httpClient.Options) *BarClient {
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
func (c *BarClient) ArgNotStruct(r *argNotStructHTTPRequest, h http.Header) (*http.Response, error) {
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
	return c.Client.Do(req)
}

// Bar calls "/bar-path" endpoint.
func (c *BarClient) Bar(r *barHTTPRequest, h http.Header) (*http.Response, error) {
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
	return c.Client.Do(req)
}

// MissingArg calls "/missing-arg-path" endpoint.
func (c *BarClient) MissingArg(h http.Header) (*http.Response, error) {
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
	return c.Client.Do(req)
}

// NoRequest calls "/no-request-path" endpoint.
func (c *BarClient) NoRequest(h http.Header) (*http.Response, error) {
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
	return c.Client.Do(req)
}

// TooManyArgs calls "/too-many-args-path" endpoint.
func (c *BarClient) TooManyArgs(r *tooManyArgsHTTPRequest, h http.Header) (*http.Response, error) {
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
	return c.Client.Do(req)
}

package bar

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/uber/zanzibar/examples/example-gateway/gen-code/uber/zanzibar/clients/foo/foo"
	"github.com/uber/zanzibar/lib/http_client"
)

// FooClient is the http client for service Foo.
type FooClient httpClient.HTTPClient

// NewClient returns a new http client for service Foo.
func NewClient(opts *httpClient.Options) *Client {
	baseURL := "http://" + opts.IP + ":" + strconv.Itoa(int(opts.Port))
	return &Client{
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

// argNotStructHTTPRequest is the http body type for endpoint argNotStruct.
type argNotStructHTTPRequest struct {
	Request string
}

// ArgNotStruct calls "/arg-not-struct-path" endpoint.
func (c *FooClient) ArgNotStruct(r *argNotStructHTTPRequest, h http.Header) (*http.Response, error) {
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

// barHTTPRequest is the http body type for endpoint bar.
type barHTTPRequest struct {
	Request BarRequest
}

// Bar calls "/bar-path" endpoint.
func (c *FooClient) Bar(r *barHTTPRequest, h http.Header) (*http.Response, error) {
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
func (c *FooClient) MissingArg(h http.Header) (*http.Response, error) {
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
func (c *FooClient) NoRequest(h http.Header) (*http.Response, error) {
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

// tooManyArgsHTTPRequest is the http body type for endpoint tooManyArgs.
type tooManyArgsHTTPRequest struct {
	Request BarRequest
	Foo     foo.FooStruct
}

// TooManyArgs calls "/too-many-args-path" endpoint.
func (c *FooClient) TooManyArgs(r *tooManyArgsHTTPRequest, h http.Header) (*http.Response, error) {
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

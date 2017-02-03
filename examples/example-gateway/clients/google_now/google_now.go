package googleNow

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/uber/zanzibar/lib/http_client"
)

// Client is the http client for googleNow.
type Client httpClient.HTTPClient

// AddCredential calls "/add-credentials" endpoint.
func (c *Client) AddCredential(r *AddCredentialRequest, h http.Header) (*http.Response, error) {
	// Generate full URL.
	// TODO(zw): this URL never change, should not be allocated every time in generated code.
	fullURL := c.BaseURL + "/add-credentials"

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

// Options to create a new googleNow client
type Options struct {
	IP   string
	Port int32
}

// NewClient returns a new http client for googleNow.
func NewClient(opts *Options) *Client {
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

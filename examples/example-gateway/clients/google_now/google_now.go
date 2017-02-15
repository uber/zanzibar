package googleNow

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/uber/zanzibar/runtime"
)

// Client is the http client for googleNow.
type Client zanzibar.HTTPClient

// AddCredential calls "/add-credentials" endpoint.
func (c *Client) AddCredential(ctx context.Context, r *AddCredentialRequest, h http.Header) (*http.Response, error) {
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
	return c.Client.Do(req.WithContext(ctx))
}

// NewClient returns a new http client for googleNow.
func NewClient(config *zanzibar.StaticConfig) *Client {
	ip := config.MustGetString("clients.googleNow.ip")
	port := config.MustGetInt("clients.googleNow.port")

	baseURL := "http://" + ip + ":" + strconv.Itoa(int(port))
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

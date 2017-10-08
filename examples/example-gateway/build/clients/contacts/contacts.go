// Code generated by zanzibar
// @generated

// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package contactsClient

import (
	"context"
	"fmt"
	"time"

	"github.com/uber/zanzibar/runtime"

	module "github.com/uber/zanzibar/examples/example-gateway/build/clients/contacts/module"
	clientsContactsContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/contacts/contacts"
)

// Client defines contacts client interface.
type Client interface {
	HTTPClient() *zanzibar.HTTPClient
	SaveContacts(
		ctx context.Context,
		reqHeaders map[string]string,
		args *clientsContactsContacts.SaveContactsRequest,
	) (*clientsContactsContacts.SaveContactsResponse, map[string]string, error)
}

// contactsClient is the http client.
type contactsClient struct {
	clientID   string
	httpClient *zanzibar.HTTPClient
}

// NewClient returns a new http client.
func NewClient(deps *module.Dependencies) Client {
	ip := deps.Default.Config.MustGetString("clients.contacts.ip")
	port := deps.Default.Config.MustGetInt("clients.contacts.port")
	baseURL := fmt.Sprintf("http://%s:%d", ip, port)
	timeout := time.Duration(deps.Default.Config.MustGetInt("clients.contacts.timeout")) * time.Millisecond

	return &contactsClient{
		clientID: "contacts",
		httpClient: zanzibar.NewHTTPClient(
			deps.Default.Logger, deps.Default.Scope,
			"contacts",
			[]string{
				"SaveContacts",
			},
			baseURL,
			timeout,
		),
	}
}

// HTTPClient returns the underlying HTTP client, should only be
// used for internal testing.
func (c *contactsClient) HTTPClient() *zanzibar.HTTPClient {
	return c.httpClient
}

// SaveContacts calls "/:userUUID/contacts" endpoint.
func (c *contactsClient) SaveContacts(
	ctx context.Context,
	headers map[string]string,
	r *clientsContactsContacts.SaveContactsRequest,
) (*clientsContactsContacts.SaveContactsResponse, map[string]string, error) {
	var defaultRes *clientsContactsContacts.SaveContactsResponse
	req := zanzibar.NewClientHTTPRequest(c.clientID, "SaveContacts", c.httpClient)

	// Generate full URL.
	fullURL := c.httpClient.BaseURL + "/" + string(r.UserUUID) + "/contacts"

	err := req.WriteJSON("POST", fullURL, headers, r)
	if err != nil {
		return defaultRes, nil, err
	}

	res, err := req.Do(ctx)
	if err != nil {
		return defaultRes, nil, err
	}

	respHeaders := map[string]string{}
	for k := range res.Header {
		respHeaders[k] = res.Header.Get(k)
	}

	res.CheckOKResponse([]int{202})

	switch res.StatusCode {
	case 202:
		var responseBody clientsContactsContacts.SaveContactsResponse
		err = res.ReadAndUnmarshalBody(&responseBody)
		if err != nil {
			return defaultRes, respHeaders, err
		}

		return &responseBody, respHeaders, nil
	}

	return defaultRes, respHeaders, &zanzibar.UnexpectedHTTPError{
		StatusCode: res.StatusCode,
		RawBody:    res.GetRawBody(),
	}
}

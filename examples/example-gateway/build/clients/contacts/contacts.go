package contactsClient

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	clientsContactsContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/contacts/contacts"
	"github.com/uber/zanzibar/runtime"
)

// ContactsClient is the http client for service Contacts.
type ContactsClient struct {
	ClientID   string
	HTTPClient *zanzibar.HTTPClient
}

// NewClient returns a new http client for service Contacts.
func NewClient(
	gateway *zanzibar.Gateway,
) *ContactsClient {
	ip := gateway.Config.MustGetString("clients.contacts.ip")
	port := gateway.Config.MustGetInt("clients.contacts.port")

	baseURL := "http://" + ip + ":" + strconv.Itoa(int(port))
	return &ContactsClient{
		ClientID:   "contacts",
		HTTPClient: zanzibar.NewHTTPClient(gateway, baseURL),
	}
}

// SaveContacts calls "/:userUUID/contacts" endpoint.
func (c *ContactsClient) SaveContacts(
	ctx context.Context,
	headers map[string]string,
	r *clientsContactsContacts.SaveContactsRequest,
) (*clientsContactsContacts.SaveContactsResponse, map[string]string, error) {

	req := zanzibar.NewClientHTTPRequest(
		c.ClientID, "saveContacts", c.HTTPClient,
	)

	// Generate full URL.
	fullURL := c.HTTPClient.BaseURL + "/" + string(r.UserUUID) + "/contacts"

	err := req.WriteJSON("POST", fullURL, headers, r)
	if err != nil {
		return nil, nil, err
	}
	res, err := req.Do(ctx)
	if err != nil {
		return nil, nil, err
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
			return nil, respHeaders, err
		}

		return &responseBody, respHeaders, nil
	}

	return nil, respHeaders, errors.Errorf(
		"Unexpected http client response (%d)", res.StatusCode,
	)
}

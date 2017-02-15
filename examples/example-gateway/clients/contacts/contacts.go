package contactsClient

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/uber/zanzibar/runtime"
)

// ContactsClient is the client itself
type ContactsClient struct {
	httpClient *http.Client
	baseURL    string
}

// SaveContactsResponse response object
type SaveContactsResponse struct {
	Res *http.Response
}

// SaveContacts will call POST /:uuid/contacts
func (contacts *ContactsClient) SaveContacts(ctx context.Context, save *SaveContactsRequest) (*SaveContactsResponse, error) {
	fullURL := contacts.baseURL + "/" + save.UserUUID + "/contacts"
	method := "POST"

	rawBody, err := save.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, fullURL, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}

	res, err := contacts.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &SaveContactsResponse{
		Res: res,
	}, nil
}

// Create makes a contacts client
func Create(config *zanzibar.StaticConfig) *ContactsClient {
	ip := config.MustGetString("clients.contacts.ip")
	port := config.MustGetInt("clients.contacts.port")

	baseURL := "http://" + ip + ":" + strconv.Itoa(int(port))
	client := &ContactsClient{
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
		},
		baseURL: baseURL,
	}
	return client
}

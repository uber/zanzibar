package contacts_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	clientContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/contacts/contacts"
	endpointContacts "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/contacts/contacts"
	"github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

func TestSaveContactsCall(t *testing.T) {
	ms := examplegatewayServiceGenerated.MustCreateTestService()
	ms.Start()
	defer ms.Stop()

	endpointReqeust := &endpointContacts.SaveContactsRequest{
		Contacts: []*endpointContacts.Contact{},
	}
	rawBody, _ := endpointReqeust.MarshalJSON()

	clientRequest := &clientContacts.SaveContactsRequest{
		UserUUID: "foo",
	}
	clientResponse := &clientContacts.SaveContactsResponse{}

	ms.MockClientNodes().Contacts.On("SaveContacts", mock.Anything, mock.Anything, clientRequest).
		Return(clientResponse, nil, nil)

	res, err := ms.MakeHTTPRequest(
		"POST", "/contacts/foo/contacts", nil, bytes.NewReader(rawBody),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "202 Accepted", res.Status)
}
